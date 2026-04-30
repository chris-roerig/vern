package cmd

import (
	"fmt"
	"os"

	"github.com/chris-roerig/vern/internal/config"
	"github.com/chris-roerig/vern/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [language] [version]",
	Short: "Create a .vern file in the current directory",
	Long: `Create a .vern file for per-project version pinning.
If language and version are provided, writes them to the file.
Otherwise creates an empty .vern file for you to edit.`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		content := ""
		if len(args) >= 1 {
			lang := args[0]
			if !config.IsValidLangName(lang) {
				ui.Error("Invalid language name: %s", lang)
				os.Exit(1)
			}
			version := ""
			if len(args) == 2 {
				version = args[1]
				if !config.IsValidVersion(version) {
					ui.Error("Invalid version: %s (expected format: M.m.p)", version)
					os.Exit(1)
				}
			}
			if version != "" {
				content = fmt.Sprintf("%s %s\n", lang, version)
			} else {
				content = fmt.Sprintf("%s\n", lang)
			}
		}

		if err := os.WriteFile(".vern", []byte(content), 0644); err != nil {
			ui.Error("Error creating .vern file: %v", err)
			os.Exit(1)
		}

		if content == "" {
			ui.Success("Created empty .vern file")
		} else {
			ui.Success("Created .vern file: %s", content)
		}
	},
}
