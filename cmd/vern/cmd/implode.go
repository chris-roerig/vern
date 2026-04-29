package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chris/vern/internal/config"
	"github.com/chris/vern/internal/ui"
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

		fmt.Println("This will remove:")
		fmt.Printf("  Binary:  %s\n", exePath)
		fmt.Printf("  Config:  %s\n", configDir)
		fmt.Printf("  Data:    %s\n", dataDir)
		fmt.Print("\nAre you sure? [y/N]: ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if scanner.Text() != "y" && scanner.Text() != "Y" {
			fmt.Println("Cancelled.")
			return
		}

		os.RemoveAll(configDir)
		os.RemoveAll(dataDir)

		if err := os.Remove(exePath); err != nil {
			ui.Error("Failed to remove binary: %v", err)
			ui.Error("You may need to run with sudo or remove it manually.")
			os.Exit(1)
		}

		ui.Success("vern has been removed. Goodbye!")

		// Check shell configs for PATH entries that should be cleaned up
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
