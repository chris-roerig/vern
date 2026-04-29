package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/chris/vern/internal/install"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up shims for version switching",
	Long:  `Create shim scripts in the vern data directory for all supported languages.`,
	Run: func(cmd *cobra.Command, args []string) {
		detectVersionManagers()

		if err := install.CreateShims(); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating shims: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Shims created successfully.")

		if !install.IsPathSet() {
			fmt.Println("\nAdd shims to your PATH:")
			fmt.Println(install.GetShellHook())
		}
	},
}

type versionManager struct {
	name     string
	envVar   string
	dir      string
	bin      string
	langs    string
}

func detectVersionManagers() {
	home, _ := os.UserHomeDir()

	managers := []versionManager{
		{"pyenv", "PYENV_ROOT", filepath.Join(home, ".pyenv"), "pyenv", "Python"},
		{"rbenv", "", filepath.Join(home, ".rbenv"), "rbenv", "Ruby"},
		{"nvm", "NVM_DIR", filepath.Join(home, ".nvm"), "", "Node.js"},
		{"fnm", "", "", "fnm", "Node.js"},
		{"asdf", "ASDF_DIR", filepath.Join(home, ".asdf"), "asdf", "multiple languages"},
		{"mise", "", "", "mise", "multiple languages"},
		{"volta", "", filepath.Join(home, ".volta"), "volta", "Node.js"},
	}

	var found []versionManager
	for _, m := range managers {
		if m.envVar != "" && os.Getenv(m.envVar) != "" {
			found = append(found, m)
			continue
		}
		if m.dir != "" {
			if _, err := os.Stat(m.dir); err == nil {
				found = append(found, m)
				continue
			}
		}
		if m.bin != "" {
			if _, err := exec.LookPath(m.bin); err == nil {
				found = append(found, m)
			}
		}
	}

	if len(found) == 0 {
		return
	}

	fmt.Println("Note: other version managers detected:")
	for _, m := range found {
		fmt.Printf("  • %s (manages %s)\n", m.name, m.langs)
	}
	fmt.Println("  vern's shims will take priority if placed earlier in PATH.")
	fmt.Println()
}
