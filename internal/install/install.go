package install

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/chris/vern/internal/config"
	"github.com/chris/vern/internal/ui"
	"github.com/chris/vern/internal/version"
)

var httpClient = &http.Client{Timeout: 5 * time.Minute}

// Verbose controls whether build output is shown
var Verbose bool

type TemplateData struct {
	Version    string
	MajorMinor string
	InstallDir string
	OS         string
	Arch       string
	ArchAlt    string
	ArchGNU    string
	OSAlt      string
	RustTarget string
}

func ArchAlt(goarch string) string {
	if goarch == "amd64" {
		return "x64"
	}
	return goarch
}

func ArchGNU(goarch string) string {
	switch goarch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	}
	return goarch
}

func OsAlt(goos string) string {
	if goos == "darwin" {
		return "macos"
	}
	return goos
}

func RustTarget(goos, goarch string) string {
	arch := ArchGNU(goarch)
	switch goos {
	case "darwin":
		return arch + "-apple-darwin"
	default:
		return arch + "-unknown-" + goos + "-gnu"
	}
}

func DownloadAndInstall(lang *config.Language, versionStr string) error {
	installDir := config.LanguageInstallDir(lang.Name, versionStr)
	if _, err := os.Stat(installDir); err == nil {
		return fmt.Errorf("version %s is already installed for %s", versionStr, lang.Name)
	}

	data := TemplateData{
		Version:    versionStr,
		MajorMinor: majorMinor(versionStr),
		InstallDir: installDir,
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		ArchAlt:    ArchAlt(runtime.GOARCH),
		ArchGNU:    ArchGNU(runtime.GOARCH),
		OSAlt:      OsAlt(runtime.GOOS),
		RustTarget: RustTarget(runtime.GOOS, runtime.GOARCH),
	}

	url, err := renderTemplate(lang.Install.DownloadTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render download URL: %w", err)
	}

	ui.Info("Downloading %s %s from %s...", lang.Name, versionStr, url)

	tmpFile, err := downloadFile(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpFile)

	// Create a temp directory for extraction and possible build
	buildDir, err := os.MkdirTemp("", "vern-build-*")
	if err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}
	defer os.RemoveAll(buildDir)

	ui.Info("Extracting %s %s...", lang.Name, versionStr)

	if err := extractArchive(tmpFile, buildDir, lang.Install.ExtractType); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// If build config is specified, compile from source
	if lang.Install.BuildConfig != "" {
		ui.Info("Building %s %s from source...", lang.Name, versionStr)
		
		// Find the extracted source directory
		sourceDir := buildDir
		entries, err := os.ReadDir(buildDir)
		if err != nil {
			return fmt.Errorf("failed to read build directory: %w", err)
		}
		if len(entries) == 1 && entries[0].IsDir() {
			sourceDir = filepath.Join(buildDir, entries[0].Name())
		}

		// Run build config
		configCmd := renderTemplateForBuild(lang.Install.BuildConfig, data)
		if Verbose { ui.Dim("Running: %s", configCmd) }
		if err := runCommand(sourceDir, configCmd); err != nil {
			return fmt.Errorf("build config failed: %w", err)
		}

		// Run build command
		if lang.Install.BuildCommand != "" {
			buildCmd := renderTemplateForBuild(lang.Install.BuildCommand, data)
			if Verbose { ui.Dim("Running: %s", buildCmd) }
			if err := runCommand(sourceDir, buildCmd); err != nil {
				return fmt.Errorf("build failed: %w", err)
			}
		}
	}

	// Create install directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Move built files to install directory
	if err := moveDirContents(buildDir, installDir); err != nil {
		os.RemoveAll(installDir)
		return fmt.Errorf("failed to move files to install directory: %w", err)
	}

	binPath := filepath.Join(installDir, lang.Install.BinRelPath)
	if _, err := os.Stat(binPath); err != nil {
		os.RemoveAll(installDir)
		return fmt.Errorf("binary not found at %s after install", lang.Install.BinRelPath)
	}

	ui.Success("Successfully installed %s %s", lang.Name, versionStr)
	return nil
}

func majorMinor(ver string) string {
	vi, err := version.ParseVersion(ver)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d.%d", vi.Major, vi.Minor)
}

func renderTemplate(tmpl string, data TemplateData) (string, error) {
	result := tmpl
	result = strings.ReplaceAll(result, "{{.Version}}", data.Version)
	result = strings.ReplaceAll(result, "{{.MajorMinor}}", data.MajorMinor)
	result = strings.ReplaceAll(result, "{{.OS}}", data.OS)
	result = strings.ReplaceAll(result, "{{.Arch}}", data.Arch)
	result = strings.ReplaceAll(result, "{{.ArchAlt}}", data.ArchAlt)
	result = strings.ReplaceAll(result, "{{.ArchGNU}}", data.ArchGNU)
	result = strings.ReplaceAll(result, "{{.OSAlt}}", data.OSAlt)
	result = strings.ReplaceAll(result, "{{.RustTarget}}", data.RustTarget)
	return result, nil
}

func downloadFile(url string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "vern-download-*")
	if err != nil {
		return "", err
	}

	reader := io.LimitReader(resp.Body, 500*1024*1024)
	if resp.ContentLength > 0 {
		reader = &progressReader{reader: reader, total: resp.ContentLength}
	}

	_, err = io.Copy(tmpFile, reader)
	if resp.ContentLength > 0 {
		fmt.Print("\n") // newline after progress
	}
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", err
	}

	tmpFile.Close()
	return tmpFile.Name(), nil
}

func extractArchive(filePath, destDir, extractType string) error {
	switch extractType {
	case "tar.gz":
		return extractTarGz(filePath, destDir)
	case "tar.xz":
		return extractTarXz(filePath, destDir)
	default:
		return fmt.Errorf("unsupported archive type: %s", extractType)
	}
}

func extractTarGz(filePath, destDir string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	return extractTar(tr, destDir)
}

func extractTarXz(filePath, destDir string) error {
	cmd := exec.Command("xz", "-dc", filePath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create xz pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start xz: %w", err)
	}

	tr := tar.NewReader(stdout)
	tarErr := extractTar(tr, destDir)

	if waitErr := cmd.Wait(); waitErr != nil {
		if tarErr != nil {
			return fmt.Errorf("xz failed: %w (tar error: %v)", waitErr, tarErr)
		}
		return fmt.Errorf("xz failed: %w", waitErr)
	}
	return tarErr
}

func extractTar(tr *tar.Reader, destDir string) error {
	// Strip top-level directory if present
	prefix := ""
	first := true

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Detect common prefix (top-level dir)
		if first {
			parts := strings.SplitN(hdr.Name, "/", 2)
			if len(parts) == 2 {
				prefix = parts[0] + "/"
			}
			first = false
		}

	// Strip prefix
		name := hdr.Name
		if prefix != "" && strings.HasPrefix(name, prefix) {
			name = strings.TrimPrefix(name, prefix)
		}
		if name == "" {
			continue
		}

		target := filepath.Join(destDir, name)

		// Path traversal protection
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("tar entry attempts path traversal: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			wf, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(wf, tr); err != nil {
				wf.Close()
				return err
			}
			wf.Close()
			os.Chmod(target, os.FileMode(hdr.Mode))
		}
	}

	return nil
}

func GetInstalledVersions(langName string) ([]string, error) {
	langDir := filepath.Join(config.InstallsDir(), langName)
	if _, err := os.Stat(langDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(langDir)
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		vi, ei := version.ParseVersion(versions[i])
		vj, ej := version.ParseVersion(versions[j])
		if ei != nil || ej != nil {
			return versions[i] < versions[j]
		}
		return vi.Compare(vj) < 0
	})

	return versions, nil
}

func GetInstalledLanguages() (map[string][]string, error) {
	result := make(map[string][]string)
	installsDir := config.InstallsDir()
	if _, err := os.Stat(installsDir); os.IsNotExist(err) {
		return result, nil
	}

	entries, err := os.ReadDir(installsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			versions, _ := GetInstalledVersions(entry.Name())
			if len(versions) > 0 {
				result[entry.Name()] = versions
			}
		}
	}
	return result, nil
}

func RemoveVersion(langName, version string) error {
	dir := config.LanguageInstallDir(langName, version)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("version %s for %s is not installed", version, langName)
	}
	return os.RemoveAll(dir)
}

func SaveVernFile(lang, version string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, ".vern")
	return os.WriteFile(path, []byte(fmt.Sprintf("%s %s\n", lang, version)), 0644)
}

func LoadVernFile() (string, string, error) {
	path, err := config.FindLocalVernFile()
	if err != nil {
		return "", "", err
	}
	return config.ParseVernFile(path)
}

func ResolveVersionForLanguage(lang *config.Language) (string, error) {
	langName, version, err := LoadVernFile()
	if err == nil && langName == lang.Name {
		return version, nil
	}

	defaults, err := config.LoadDefaults()
	if err != nil {
		return "", fmt.Errorf("no version set for %s", lang.Name)
	}

	v, ok := defaults[lang.Name]
	if !ok {
		return "", fmt.Errorf("no version set for %s", lang.Name)
	}
	return v, nil
}

func GetShimScript(lang *config.Language) string {
	return fmt.Sprintf(`#!/bin/sh
exec "%s" "$@"
`, filepath.Join(config.DataDir(), "versions", lang.Name, "current", lang.BinaryName))
}

func CreateShims() error {
	shimsDir := config.ShimsDir()
	if err := os.MkdirAll(shimsDir, 0755); err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	for _, lang := range cfg.Languages {
		shimPath := filepath.Join(shimsDir, lang.BinaryName)
		script := fmt.Sprintf(`#!/bin/sh
VERN_DATA="%s"
LANG="%s"
BIN="%s"

# Check for .vern file
dir="$PWD"
while [ "$dir" != "/" ]; do
    if [ -f "$dir/.vern" ]; then
        vern_file=$(cat "$dir/.vern")
        if echo "$vern_file" | grep -q "^$LANG "; then
            version=$(echo "$vern_file" | cut -d' ' -f2)
            exec "$VERN_DATA/installs/$LANG/$version/$BIN" "$@"
        fi
    fi
    dir=$(dirname "$dir")
done

# Check defaults
if [ -f "$VERN_DATA/defaults.yaml" ]; then
    version=$(grep "^$LANG:" "$VERN_DATA/defaults.yaml" | cut -d: -f2 | tr -d ' ')
    if [ -n "$version" ]; then
        exec "$VERN_DATA/installs/$LANG/$version/$BIN" "$@"
    fi
fi

echo "No version set for $LANG"
exit 1
`, config.DataDir(), lang.Name, lang.Install.BinRelPath)

		if err := os.WriteFile(shimPath, []byte(script), 0755); err != nil {
			return err
		}
	}
	return nil
}

func IsPathSet() bool {
	path := os.Getenv("PATH")
	shimsDir := config.ShimsDir()
	return strings.Contains(path, shimsDir)
}

func renderTemplateForBuild(templateStr string, data TemplateData) string {
	result := strings.ReplaceAll(templateStr, "{{.InstallDir}}", data.InstallDir)
	result = strings.ReplaceAll(result, "{{.Version}}", data.Version)
	result = strings.ReplaceAll(result, "{{.MajorMinor}}", data.MajorMinor)
	return result
}

func runCommand(dir, cmdStr string) error {
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Dir = dir
	if Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

type progressReader struct {
	reader  io.Reader
	total   int64
	current int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.current += int64(n)
	pct := float64(pr.current) / float64(pr.total) * 100
	mb := float64(pr.current) / 1024 / 1024
	totalMB := float64(pr.total) / 1024 / 1024
	fmt.Fprintf(os.Stdout, "\r  %.1f/%.1f MB (%.0f%%)", mb, totalMB, pct)
	return n, err
}

func moveDirContents(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if err := os.Rename(srcPath, dstPath); err != nil {
			// If rename fails (cross-device), try copy
			if err := copyDirRecursive(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyDirRecursive(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcInfo.IsDir() {
		return copyFile(src, dst)
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if err := copyDirRecursive(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func GetShellHook() string {
	shimsDir := config.ShimsDir()
	return fmt.Sprintf(`# Add vern shims to PATH
export PATH="%s:$PATH"`, shimsDir)
}
