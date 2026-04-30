package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/template"

	"github.com/chris-roerig/vern/internal/config"
	"github.com/chris-roerig/vern/internal/ui"
	"github.com/chris-roerig/vern/internal/version"
)

// TemplateData holds variables available in download URL and build command templates.
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

// DownloadAndInstall downloads, extracts, and optionally builds a language version.
func DownloadAndInstall(lang *config.Language, versionStr string, opts Options) error {
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

	tmpFile, err := DownloadFile(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpFile)

	buildDir, err := os.MkdirTemp("", "vern-build-*")
	if err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}
	defer os.RemoveAll(buildDir)

	ui.Info("Extracting %s %s...", lang.Name, versionStr)

	if err := ExtractArchive(tmpFile, buildDir, lang.Install.ExtractType); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	if lang.Install.BuildConfig != "" {
		if err := buildFromSource(lang, data, buildDir, opts); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

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

func buildFromSource(lang *config.Language, data TemplateData, buildDir string, opts Options) error {
	ui.Info("Building %s %s from source...", lang.Name, data.Version)

	sourceDir := buildDir
	entries, err := os.ReadDir(buildDir)
	if err != nil {
		return fmt.Errorf("failed to read build directory: %w", err)
	}
	if len(entries) == 1 && entries[0].IsDir() {
		sourceDir = filepath.Join(buildDir, entries[0].Name())
	}

	configCmd := renderBuildTemplate(lang.Install.BuildConfig, data)
	if opts.Verbose {
		ui.Dim("Running: %s", configCmd)
	}
	if err := runCommand(sourceDir, configCmd, opts); err != nil {
		if !opts.Verbose {
			ui.Warn("Hint: run with --verbose to see build output")
		}
		return fmt.Errorf("build config failed: %w", err)
	}

	if lang.Install.BuildCommand != "" {
		buildCmd := renderBuildTemplate(lang.Install.BuildCommand, data)
		if opts.Verbose {
			ui.Dim("Running: %s", buildCmd)
		}
		if err := runCommand(sourceDir, buildCmd, opts); err != nil {
			if !opts.Verbose {
				ui.Warn("Hint: run with --verbose to see build output")
			}
			return fmt.Errorf("build failed: %w", err)
		}
	}
	return nil
}

func majorMinor(ver string) string {
	vi, err := version.ParseVersion(ver)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d.%d", vi.Major, vi.Minor)
}

// renderTemplate renders a Go text/template string with TemplateData.
func renderTemplate(tmpl string, data TemplateData) (string, error) {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// renderBuildTemplate renders build commands, which use the same template syntax.
func renderBuildTemplate(tmpl string, data TemplateData) string {
	result, err := renderTemplate(tmpl, data)
	if err != nil {
		// Fallback to simple replacement for backward compatibility
		result = strings.ReplaceAll(tmpl, "{{.InstallDir}}", data.InstallDir)
		result = strings.ReplaceAll(result, "{{.Version}}", data.Version)
		result = strings.ReplaceAll(result, "{{.MajorMinor}}", data.MajorMinor)
	}
	return result
}

// GetInstalledVersions returns installed versions for a language, sorted by semver.
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

// GetInstalledLanguages returns a map of language name to installed versions.
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
			versions, err := GetInstalledVersions(entry.Name())
			if err != nil {
				continue
			}
			if len(versions) > 0 {
				result[entry.Name()] = versions
			}
		}
	}
	return result, nil
}

// RemoveVersion removes an installed language version.
func RemoveVersion(langName, ver string) error {
	dir := config.LanguageInstallDir(langName, ver)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("version %s for %s is not installed", ver, langName)
	}
	return os.RemoveAll(dir)
}

// LoadVernFile finds and parses the nearest .vern file.
func LoadVernFile() (string, string, error) {
	path, err := config.FindLocalVernFile()
	if err != nil {
		return "", "", err
	}
	return config.ParseVernFile(path)
}

// ResolveVersionForLanguage resolves the active version for a language
// by checking .vern files and then global defaults.
func ResolveVersionForLanguage(lang *config.Language) (string, error) {
	langName, ver, err := LoadVernFile()
	if err == nil && langName == lang.Name {
		return ver, nil
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

func moveDirContents(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if err := os.Rename(srcPath, dstPath); err != nil {
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
		if err := copyDirRecursive(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name())); err != nil {
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
