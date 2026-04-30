package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chris-roerig/vern/internal/config"
	"github.com/chris-roerig/vern/internal/install"
	"github.com/chris-roerig/vern/internal/ui"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <language>",
	Short: "Remove installed language versions",
	Long: `Interactively select and remove installed versions of a language.
Uses checkbox multi-select - space to select, enter to confirm.`,
	Args: cobra.ExactArgs(1),
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

		type versionItem struct {
			name    string
			version string
			selected bool
		}

		items := make([]versionItem, len(versions))
		for i, v := range versions {
			marker := "  "
			if v == currentDefault {
				marker = "* "
			}
			items[i] = versionItem{name: fmt.Sprintf("%s%s", marker, v), version: v}
		}

		fmt.Println("Select versions to remove (comma-separated numbers, e.g., 1,3,4):")
		for i, item := range items {
			fmt.Printf("  %d) %s\n", i+1, item.name)
		}
		fmt.Print("Selection: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()

		if input == "" {
			fmt.Println("No versions selected")
			return
		}

		var selectedItems []versionItem
		parts := strings.Split(input, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			var idx int
			fmt.Sscanf(p, "%d", &idx)
			if idx >= 1 && idx <= len(items) {
				selectedItems = append(selectedItems, items[idx-1])
			}
		}

		if len(selectedItems) == 0 {
			fmt.Println("No valid selections")
			return
		}

		fmt.Printf("Remove %d version(s): ", len(selectedItems))
		for _, item := range selectedItems {
			fmt.Printf("%s ", item.version)
		}
		fmt.Print("\nConfirm? [y/N]: ")
		scanner.Scan()
		confirm := scanner.Text()
		if strings.ToLower(confirm) != "y" {
			fmt.Println("Cancelled")
			return
		}

		removedDefault := false
		for _, item := range selectedItems {
			if err := install.RemoveVersion(langName, item.version); err != nil {
				ui.Error("Error removing %s: %v", item.version, err)
			} else {
				ui.Success("Removed %s %s", langName, item.version)
				if item.version == currentDefault {
					removedDefault = true
				}
			}
		}

		if removedDefault {
			remaining, _ := install.GetInstalledVersions(langName)
			if len(remaining) > 0 {
				newDefault := remaining[len(remaining)-1]
				defaults[langName] = newDefault
				config.SaveDefaults(defaults)
				ui.Success("Set %s %s as new default", langName, newDefault)
			} else {
				delete(defaults, langName)
				config.SaveDefaults(defaults)
				ui.Info("No versions remaining for %s", langName)
			}
		}
	},
}
