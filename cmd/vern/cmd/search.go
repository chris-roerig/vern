package cmd

import (
	"fmt"
	"os"

	"github.com/chris-roerig/vern/internal/config"
	"github.com/chris-roerig/vern/internal/ui"
	"github.com/chris-roerig/vern/internal/version"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <language> [version]",
	Short: "Search available versions for a language",
	Long: `List available versions for a language. Optionally filter by partial version.
Examples: vern search python, vern search python 3.12, vern search go 1`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		langName := args[0]
		filter := ""
		if len(args) > 1 {
			filter = args[1]
		}

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

		ui.Info("Fetching versions for %s...", langName)
		available, err := version.FetchAvailableVersions(lang)
		if err != nil {
			ui.Error("Error: %v", err)
			os.Exit(1)
		}

		var filtered []version.VersionInfo
		if filter == "" {
			filtered = available
		} else {
			for _, v := range available {
				if matchesFilter(v, filter) {
					filtered = append(filtered, v)
				}
			}
		}

		if len(filtered) == 0 {
			ui.Warn("No versions found matching %q", filter)
			return
		}

		for _, v := range filtered {
			fmt.Println(v.Full)
		}
		ui.Dim("%d version(s)", len(filtered))
	},
}

func matchesFilter(v version.VersionInfo, filter string) bool {
	// Try exact prefix match on full version string
	if len(filter) <= len(v.Full) && v.Full[:len(filter)] == filter {
		return true
	}
	return false
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
