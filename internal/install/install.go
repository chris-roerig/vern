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
	"strings"

	"github.com/chris/vern/internal/config"
	"github.com/chris/vern/internal/version"
)

type TemplateData struct {
	Version    string
	MajorMinor string
}

func DownloadAndInstall(lang *config.Language, versionStr string) error {
	installDir := config.LanguageInstallDir(lang.Name, versionStr)
	if _, err := os.Stat(installDir); err == nil {
		return fmt.Errorf("version %s is already installed for %s", versionStr, lang.Name)
	}

	data := TemplateData{
		Version:    versionStr,
		MajorMinor: majorMinor(versionStr),
	}

	url, err := renderTemplate(lang.Install.DownloadTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render download URL: %w", err)
	}

	fmt.Printf("Downloading %s %s from %s...\n", lang.Name, versionStr, url)

	tmpFile, err := downloadFile(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpFile)

	fmt.Printf("Extracting %s %s...\n", lang.Name, versionStr)

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	if err := extractArchive(tmpFile, installDir, lang.Install.ExtractType); err != nil {
		os.RemoveAll(installDir)
		return fmt.Errorf("extraction failed: %w", err)
	}

	binPath := filepath.Join(installDir, lang.Install.BinRelPath)
	if _, err := os.Stat(binPath); err != nil {
		os.RemoveAll(installDir)
		return fmt.Errorf("binary not found at %s after install", lang.Install.BinRelPath)
	}

	fmt.Printf("Successfully installed %s %s\n", lang.Name, versionStr)
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
	return result, nil
}

func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
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

	_, err = io.Copy(tmpFile, resp.Body)
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
	defer cmd.Wait()

	tr := tar.NewReader(stdout)
	return extractTar(tr, destDir)
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
VERN_SHIMS="%s"
VERN_DATA="%s"

vern_resolve() {
    # Check for .vern file
    dir="$PWD"
    while [ "$dir" != "/" ]; do
        if [ -f "$dir/.vern" ]; then
            cat "$dir/.vern"
            return
        fi
        dir=$(dirname "$dir")
    done

    # Check defaults
    if [ -f "$VERN_DATA/defaults.yaml" ]; then
        grep "^%s:" "$VERN_DATA/defaults.yaml" | cut -d: -f2 | tr -d ' '
    fi
}

lang_ver=$(vern_resolve)
if [ -z "$lang_ver" ]; then
    echo "No version set for %s"
    exit 1
fi

        exec "$VERN_DATA/installs/%s/$lang_ver/%s" "$@"
`, shimsDir, config.DataDir(), lang.Name, lang.Name, lang.Name, lang.Install.BinRelPath)

		if err := os.WriteFile(shimPath, []byte(script), 0755); err != nil {
			return err
		}
	}
	return nil
}

func GetShellHook() string {
	shimsDir := config.ShimsDir()
	return fmt.Sprintf(`# Add vern shims to PATH
export PATH="%s:$PATH"`, shimsDir)
}
