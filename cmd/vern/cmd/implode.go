package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chris/vern/internal/config"
	"github.com/chris/vern/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var implodeCmd = &cobra.Command{
	Use:   "implode",
	Short: "Uninstall vern and all installed languages",
	Long:  `Remove the vern binary, all installed language versions, shims, and configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		exePath, err := os.Executable()
		if err != nil {
			ui.Error("Error: %v", err)
			os.Exit(1)
		}

		configDir := config.ConfigDir()
		dataDir := config.DataDir()
		installsDir := config.InstallsDir()

		options := []string{
			"Binary only        - just the vern binary",
			"Config             - languages.yaml, defaults, shims",
			"All languages      - all installed language versions",
			"Everything         - binary, config, languages, all data",
		}

		prompt := promptui.Select{
			Label: "What would you like to remove",
			Items: options,
			Size:  len(options),
		}

		idx, _, err := prompt.Run()
		if err != nil {
			fmt.Println("Cancelled.")
			return
		}

		removeBin, removeConfig, removeLangs := false, false, false
		switch idx {
		case 0:
			removeBin = true
		case 1:
			removeConfig = true
		case 2:
			removeLangs = true
		case 3:
			removeBin, removeConfig, removeLangs = true, true, true
		}

		fmt.Println("\nThis will remove:")
		if removeBin {
			fmt.Printf("  Binary:     %s\n", exePath)
		}
		if removeConfig {
			fmt.Printf("  Config:     %s\n", configDir)
		}
		if removeLangs {
			fmt.Printf("  Languages:  %s\n", installsDir)
		}
		if removeConfig && removeLangs {
			fmt.Printf("  Data:       %s\n", dataDir)
		}

		fmt.Print("\nConfirm? [y/N]: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if scanner.Text() != "y" && scanner.Text() != "Y" {
			fmt.Println("Cancelled.")
			return
		}

		if removeLangs {
			if removeConfig {
				os.RemoveAll(dataDir)
				ui.Success("Removed all data: %s", dataDir)
			} else {
				os.RemoveAll(installsDir)
				ui.Success("Removed all installed languages: %s", installsDir)
			}
		}

		if removeConfig && !removeLangs {
			os.RemoveAll(configDir)
			os.RemoveAll(config.ShimsDir())
			os.Remove(config.DefaultsPath())
			ui.Success("Removed config and shims")
		} else if removeConfig {
			os.RemoveAll(configDir)
			ui.Success("Removed config: %s", configDir)
		}

		if removeBin {
			if err := os.Remove(exePath); err != nil {
				ui.Error("Failed to remove binary: %v", err)
				ui.Error("You may need to run with sudo or remove it manually.")
			} else {
				ui.Success("Removed binary: %s", exePath)
			}
		}

		binDir := filepath.Dir(exePath)
		shimsDir := config.ShimsDir()
		checkShellConfigs(binDir, shimsDir)
	},
}

func checkShellConfigs(binDir, shimsDir string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	rcFiles := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".config", "fish", "config.fish"),
	}

	var found []string
	for _, rc := range rcFiles {
		data, err := os.ReadFile(rc)
		if err != nil {
			continue
		}
		content := string(data)
		if strings.Contains(content, shimsDir) || strings.Contains(content, binDir) {
			found = append(found, rc)
		}
	}

	if len(found) > 0 {
		ui.Warn("You may want to remove vern PATH entries from:")
		for _, f := range found {
			fmt.Printf("  %s\n", f)
		}
	}
}
