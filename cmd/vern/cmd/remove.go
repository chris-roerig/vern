package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chris-roerig/vern/internal/config"
	"github.com/chris-roerig/vern/internal/install"
	"github.com/chris-roerig/vern/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <language>",
	Short: "Remove installed language versions",
	Long:  `Interactively select a version to remove.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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

		defaults, _ := config.LoadDefaults()
		currentDefault := defaults[langName]

		// Build display items
		items := make([]string, len(versions))
		for i, v := range versions {
			if v == currentDefault {
				items[i] = v + " (default)"
			} else {
				items[i] = v
			}
		}

		prompt := promptui.Select{
			Label: "Select version to remove",
			Items: items,
			Size:  len(items),
		}

		idx, _, err := prompt.Run()
		if err != nil {
			fmt.Println("Cancelled.")
			return
		}

		selected := versions[idx]

		fmt.Printf("Remove %s %s? [y/N]: ", langName, selected)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if strings.ToLower(scanner.Text()) != "y" {
			fmt.Println("Cancelled.")
			return
		}

		if err := install.RemoveVersion(langName, selected); err != nil {
			ui.Error("Error removing %s: %v", selected, err)
			os.Exit(1)
		}
		ui.Success("Removed %s %s", langName, selected)

		if selected == currentDefault {
			remaining, _ := install.GetInstalledVersions(langName)
			if len(remaining) > 0 {
				newDefault := remaining[len(remaining)-1]
				defaults[langName] = newDefault
				if err := config.SaveDefaults(defaults); err != nil {
					ui.Warn("Failed to save defaults: %v", err)
				}
				ui.Success("Set %s %s as new default", langName, newDefault)
			} else {
				delete(defaults, langName)
				if err := config.SaveDefaults(defaults); err != nil {
					ui.Warn("Failed to save defaults: %v", err)
				}
				ui.Info("No versions remaining for %s", langName)
			}
		}
	},
}
