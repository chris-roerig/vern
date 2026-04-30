package cmd

import (
	"fmt"
	"os"

	"github.com/chris-roerig/vern/internal/config"
	"github.com/chris-roerig/vern/internal/install"
	"github.com/chris-roerig/vern/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [language]",
	Short: "List installed versions",
	Long: `List installed language versions. If no language is specified, lists all installed languages.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		defaults, _ := config.LoadDefaults()

		if len(args) == 0 {
			installed, err := install.GetInstalledLanguages()
			if err != nil {
				ui.Error("Error: %v", err)
				os.Exit(1)
			}
			if len(installed) == 0 {
				ui.Info("No languages installed yet.")
				fmt.Println("Run 'vern install <language>' to install one.")
				return
			}
			for lang, versions := range installed {
				fmt.Printf("%s (%d versions):\n", lang, len(versions))
				for _, v := range versions {
					marker := ""
					if defaults[lang] == v {
						marker = " * (default)"
					}
					fmt.Printf("  %s%s\n", v, marker)
				}
			}
			return
		}

		langName := args[0]
		versions, err := install.GetInstalledVersions(langName)
		if err != nil {
			ui.Error("Error: %v", err)
			os.Exit(1)
		}
		if len(versions) == 0 {
			ui.Info("No versions installed for %s.", langName)
			return
		}
		fmt.Printf("%s versions:\n", langName)
		for _, v := range versions {
			marker := ""
			if defaults[langName] == v {
				marker = " * (default)"
			}
			fmt.Printf("  %s%s\n", v, marker)
		}
	},
}
