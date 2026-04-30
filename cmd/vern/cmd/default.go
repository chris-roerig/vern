package cmd

import (
	"fmt"
	"os"

	"github.com/chris-roerig/vern/internal/config"
	"github.com/chris-roerig/vern/internal/install"
	"github.com/chris-roerig/vern/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var defaultCmd = &cobra.Command{
	Use:   "default <language> [version]",
	Short: "Set or show default version for a language",
	Long: `Set the default version for a language globally.
If no version is supplied, shows available installed versions to choose from.`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		langName := args[0]

		if len(args) == 2 {
			version := args[1]
			versions, err := install.GetInstalledVersions(langName)
			if err != nil {
				ui.Error("Error: %v", err)
				os.Exit(1)
			}

			found := false
			for _, v := range versions {
				if v == version {
					found = true
					break
				}
			}
			if !found {
				if len(versions) == 0 {
					ui.Info("No versions installed for %s.", langName)
				} else {
					ui.Warn("Version %s not installed for %s.", version, langName)
					fmt.Printf("Installed versions:\n")
					for _, v := range versions {
						fmt.Printf("  %s\n", v)
					}
				}
				os.Exit(1)
			}

			defaults, _ := config.LoadDefaults()
			defaults[langName] = version
			if err := config.SaveDefaults(defaults); err != nil {
				ui.Error("Error saving defaults: %v", err)
				os.Exit(1)
			}
			ui.Success("Set %s %s as default", langName, version)
			return
		}

		versions, err := install.GetInstalledVersions(langName)
		if err != nil {
			ui.Error("Error: %v", err)
			os.Exit(1)
		}
		if len(versions) == 0 {
			ui.Info("No versions installed for %s. Run 'vern install %s' first.", langName, langName)
			os.Exit(1)
		}

		defaults, _ := config.LoadDefaults()
		currentDefault := defaults[langName]

		items := make([]string, len(versions))
		for i, v := range versions {
			marker := ""
			if v == currentDefault {
				marker = " (current)"
			}
			items[i] = fmt.Sprintf("%s%s", v, marker)
		}

		prompt := promptui.Select{
			Label: "Select default version for " + langName,
			Items: items,
			Size:  len(items),
		}

		idx, _, err := prompt.Run()
		if err != nil {
			fmt.Printf("Cancelled\n")
			return
		}

		selectedVersion := versions[idx]

		defaults[langName] = selectedVersion
		if err := config.SaveDefaults(defaults); err != nil {
			ui.Error("Error saving defaults: %v", err)
			os.Exit(1)
		}
		ui.Success("Set %s %s as default", langName, selectedVersion)
	},
}
