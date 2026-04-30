package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chris-roerig/vern/internal/config"
	"github.com/chris-roerig/vern/internal/ui"
)

// CreateShims scans installed language versions and creates shim scripts
// for every executable found in each language's bin directory.
func CreateShims() error {
	shimsDir := config.ShimsDir()
	if err := os.MkdirAll(shimsDir, 0755); err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	seen := make(map[string]string) // binary name -> language that claimed it

	for _, lang := range cfg.Languages {
		binDir := filepath.Dir(lang.Install.BinRelPath)
		if binDir == "." {
			binDir = ""
		}

		versions, _ := GetInstalledVersions(lang.Name)
		if len(versions) == 0 {
			writeShim(shimsDir, lang.Name, lang.Install.BinRelPath, seen)
			continue
		}

		latestVersion := versions[len(versions)-1]
		installDir := config.LanguageInstallDir(lang.Name, latestVersion)

		var scanDir string
		if binDir == "" {
			scanDir = installDir
		} else {
			scanDir = filepath.Join(installDir, binDir)
		}

		entries, err := os.ReadDir(scanDir)
		if err != nil {
			writeShim(shimsDir, lang.Name, lang.Install.BinRelPath, seen)
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			info, err := entry.Info()
			if err != nil || info.Mode()&0111 == 0 {
				continue
			}
			var relPath string
			if binDir == "" {
				relPath = entry.Name()
			} else {
				relPath = filepath.Join(binDir, entry.Name())
			}
			writeShim(shimsDir, lang.Name, relPath, seen)
		}
	}
	return nil
}

// writeShim creates a single shim script, warning on name collisions.
func writeShim(shimsDir, langName, binRelPath string, seen map[string]string) {
	if !config.IsValidLangName(langName) || !config.IsValidBinPath(binRelPath) {
		return
	}
	binName := filepath.Base(binRelPath)

	if prev, ok := seen[binName]; ok {
		ui.Warn("Shim collision: %s claimed by both %s and %s (using %s)", binName, prev, langName, langName)
	}
	seen[binName] = langName

	shimPath := filepath.Join(shimsDir, binName)
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
            # Validate version string to prevent path traversal
            if ! echo "$version" | grep -qE '^[0-9]+(\.[0-9]+)*$'; then
                echo "Invalid version in .vern: $version"
                exit 1
            fi
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
`, config.DataDir(), langName, binRelPath)

	if err := os.WriteFile(shimPath, []byte(script), 0755); err != nil {
		ui.Warn("Failed to write shim %s: %v", binName, err)
	}
}

// IsPathSet returns true if the vern shims directory is in PATH.
func IsPathSet() bool {
	path := os.Getenv("PATH")
	shimsDir := config.ShimsDir()
	return contains(path, shimsDir)
}

func contains(path, dir string) bool {
	for _, p := range filepath.SplitList(path) {
		if p == dir {
			return true
		}
	}
	return false
}

// GetShellHook returns the shell export line to add shims to PATH.
func GetShellHook() string {
	return fmt.Sprintf("# Add vern shims to PATH\nexport PATH=\"%s:$PATH\"", config.ShimsDir())
}
