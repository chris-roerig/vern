package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chris/vern/internal/config"
	"github.com/chris/vern/internal/install"
	"github.com/chris/vern/internal/ui"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show installed languages, versions, and disk usage",
	Run: func(cmd *cobra.Command, args []string) {
		installed, err := install.GetInstalledLanguages()
		if err != nil {
			ui.Error("Error: %v", err)
			os.Exit(1)
		}
		if len(installed) == 0 {
			ui.Info("No languages installed.")
			return
		}

		defaults, _ := config.LoadDefaults()
		var total int64

		for lang, versions := range installed {
			langDir := filepath.Join(config.InstallsDir(), lang)
			langSize := dirSize(langDir)
			total += langSize
			fmt.Printf("%s (%s)\n", lang, formatSize(langSize))
			for _, v := range versions {
				vDir := filepath.Join(langDir, v)
				vSize := dirSize(vDir)
				marker := ""
				if defaults[lang] == v {
					marker = " * (default)"
				}
				fmt.Printf("  %s  %s%s\n", v, formatSize(vSize), marker)
			}
		}

		fmt.Printf("\nTotal: %s\n", formatSize(total))
	},
}

func dirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func formatSize(bytes int64) string {
	const (
		mb = 1024 * 1024
		gb = 1024 * mb
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	default:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
}
