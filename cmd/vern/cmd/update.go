package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chris/vern/internal/config"
	"github.com/chris/vern/internal/ui"
	"github.com/spf13/cobra"
)

var httpClient = &http.Client{Timeout: 5 * time.Minute}

var (
	updateOnlySelf  bool
	updateOnlyLangs bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update vern binary and language list",
	Long: `Update vern to the latest version and/or update the supported language list.
By default, updates both. Use --only-self or --only-langs to update just one.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !updateOnlyLangs {
			if err := updateSelf(); err != nil {
				ui.Error("Error updating vern: %v", err)
			}
		}
		if !updateOnlySelf {
			if err := updateLangs(); err != nil {
				ui.Error("Error updating language list: %v", err)
			}
		}
	},
}

func updateSelf() error {
	ui.Info("Checking for vern updates...")

	resp, err := httpClient.Get("https://api.github.com/repos/chris-roerig/vern/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release info: %w", err)
	}

	latestVersion := release.TagName
	if latestVersion == "" {
		return fmt.Errorf("could not determine latest version")
	}

	// Compare versions (latestVersion has "v" prefix, Version doesn't)
	if latestVersion == "v"+Version || latestVersion == Version {
		ui.Info("Vern is already at the latest version: %s", Version)
		return nil
	}

	ui.Info("Updating vern from %s to %s...", Version, latestVersion)

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	// Use architecture names that match release assets
	// Release assets use: vern-v0.1.0-linux-amd64, vern-v0.1.0-darwin-arm64, etc.
	// Go's GOARCH: amd64→amd64, arm64→arm64 (these match)
	// Go's GOOS: linux→linux, darwin→darwin (these match)
	
	assetName := fmt.Sprintf("vern-%s-%s-%s", latestVersion, goos, goarch)
	downloadURL := fmt.Sprintf("https://github.com/chris-roerig/vern/releases/download/%s/%s", latestVersion, assetName)

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine current executable path: %w", err)
	}

	ui.Info("Downloading from %s...", downloadURL)
	tmpFile, err := downloadBinary(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpFile)

	// Verify checksum
	checksumURL := downloadURL + ".sha256"
	if err := verifyChecksum(tmpFile, checksumURL); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	if err := os.Rename(exePath, exePath+".old"); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	if err := os.Rename(tmpFile, exePath); err != nil {
		os.Rename(exePath+".old", exePath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}
	os.Remove(exePath + ".old")

	ui.Success("Updated vern to %s. Restart your shell.", latestVersion)
	return nil
}

func updateLangs() error {
	ui.Info("Checking for language list updates...")

	manifestURL := "https://raw.githubusercontent.com/chris-roerig/vern/main/languages/manifest.json"
	resp, err := httpClient.Get(manifestURL)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("manifest returned %d", resp.StatusCode)
	}

	var manifest struct {
		LatestLangsVersion string `json:"latest_langs_version"`
		LangsURL           string `json:"langs_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	currentVersion := config.LoadLangsVersion()
	if manifest.LatestLangsVersion == currentVersion {
		ui.Info("Language list is already at the latest version: %s", currentVersion)
		return nil
	}

	ui.Info("Updating language list from %s to %s...", currentVersion, manifest.LatestLangsVersion)

	langsResp, err := httpClient.Get(manifest.LangsURL)
	if err != nil {
		return fmt.Errorf("failed to download language list: %w", err)
	}
	defer langsResp.Body.Close()

	if langsResp.StatusCode != http.StatusOK {
		return fmt.Errorf("language list returned %d", langsResp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(langsResp.Body, 1*1024*1024)) // 1MB max
	if err != nil {
		return fmt.Errorf("failed to read language list: %w", err)
	}

	configPath := config.ConfigDir()
	langsPath := filepath.Join(configPath, "languages.yaml")
	if err := os.WriteFile(langsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save language list: %w", err)
	}

	config.SaveLangsVersion(manifest.LatestLangsVersion)
	ui.Success("Updated language list to %s", manifest.LatestLangsVersion)
	return nil
}

func downloadBinary(url string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "vern-update-*")
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(tmpFile, io.LimitReader(resp.Body, 100*1024*1024)); err != nil { // 100MB max
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", err
	}

	tmpFile.Close()
	os.Chmod(tmpFile.Name(), 0755)

	return tmpFile.Name(), nil
}

func verifyChecksum(filePath, checksumURL string) error {
	resp, err := httpClient.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("failed to fetch checksum: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ui.Warn("No checksum available, skipping verification")
		return nil
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		return fmt.Errorf("failed to read checksum: %w", err)
	}

	// sha256sum format: "<hash>  <filename>" or "<hash> <filename>"
	expectedHash := strings.Fields(strings.TrimSpace(string(data)))[0]

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actualHash := hex.EncodeToString(h.Sum(nil))

	if actualHash != expectedHash {
		return fmt.Errorf("expected %s, got %s", expectedHash, actualHash)
	}
	ui.Success("Checksum verified.")
	return nil
}

func init() {
	updateCmd.Flags().BoolVar(&updateOnlySelf, "only-self", false, "Only update vern binary")
	updateCmd.Flags().BoolVar(&updateOnlyLangs, "only-langs", false, "Only update language list")
}
