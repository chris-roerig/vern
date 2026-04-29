package cmd

import (
	"os"

	"github.com/chris/vern/internal/config"
	"github.com/chris/vern/internal/install"
	"github.com/chris/vern/internal/ui"
	"github.com/spf13/cobra"
)

var whichCmd = &cobra.Command{
	Use:   "which <language>",
	Short: "Show resolved version for a language",
	Long:  `Show which version of a language would be used in the current directory.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		langName := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			ui.Error("Error loading config: %v", err)
			os.Exit(1)
		}

		var lang *config.Language
		for _, l := range cfg.Languages {
			if l.Name == langName {
				lang = &l
				break
			}
		}
		if lang == nil {
			ui.Error("Unknown language: %s", langName)
			os.Exit(1)
		}

		ver, err := install.ResolveVersionForLanguage(lang)
		if err != nil {
			ui.Warn("No version set for %s", langName)
			os.Exit(1)
		}

		// Check if it's from .vern or defaults
		vernLang, vernVer, vernErr := install.LoadVernFile()
		if vernErr == nil && vernLang == langName {
			ui.Info("%s %s (from .vern)", langName, vernVer)
		} else {
			ui.Info("%s %s (default)", langName, ver)
		}
	},
}

func init() {
	rootCmd.AddCommand(whichCmd)
}
