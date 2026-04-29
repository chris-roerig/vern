package cmd

import (
	"fmt"
	"os"

	"github.com/chris/vern/internal/config"
	"github.com/chris/vern/internal/install"
	"github.com/chris/vern/internal/version"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <language> [version]",
	Short: "Install a language version",
	Long: `Install a programming language version. If no version is supplied, installs the latest.
Partial versions (e.g., "3" or "3.11") will install the latest matching version.`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		langName := args[0]
		versionArg := ""
		if len(args) > 1 {
			versionArg = args[1]
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
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
			fmt.Fprintf(os.Stderr, "Unsupported language: %s\n", langName)
			fmt.Fprintf(os.Stderr, "Supported languages: ")
			for i, l := range cfg.Languages {
				if i > 0 {
					fmt.Fprintf(os.Stderr, ", ")
				}
				fmt.Fprintf(os.Stderr, l.Name)
			}
			fmt.Fprintf(os.Stderr, "\n")
			os.Exit(1)
		}

		resolvedVersion, err := version.ResolveVersion(lang, versionArg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Installing %s %s...\n", lang.Name, resolvedVersion)

		if err := install.DownloadAndInstall(lang, resolvedVersion); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		defaults, _ := config.LoadDefaults()
		if defaults[lang.Name] == "" {
			defaults[lang.Name] = resolvedVersion
			config.SaveDefaults(defaults)
			fmt.Printf("Set %s %s as default\n", lang.Name, resolvedVersion)
		}
	},
}
