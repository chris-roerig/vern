package cmd

import (
	"fmt"
	"os"

	"github.com/chris/vern/internal/config"
	"github.com/chris/vern/internal/install"
	"github.com/chris/vern/internal/ui"
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

		verbose, _ := cmd.Flags().GetBool("verbose")
		install.Verbose = verbose

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
			ui.Error("Unsupported language: %s", langName)
			fmt.Fprintf(os.Stderr, "Supported languages: ")
			for i, l := range cfg.Languages {
				if i > 0 {
					fmt.Fprintf(os.Stderr, ", ")
				}
				fmt.Fprintf(os.Stderr, "%s", l.Name)
			}
			fmt.Fprintf(os.Stderr, "\n")
			os.Exit(1)
		}

		resolvedVersion, err := version.ResolveVersion(lang, versionArg)
		if err != nil {
			ui.Error("Error resolving version: %v", err)
			os.Exit(1)
		}

		ui.Info("Installing %s %s...", lang.Name, resolvedVersion)

		if err := install.DownloadAndInstall(lang, resolvedVersion); err != nil {
			ui.Error("Error: %v", err)
			os.Exit(1)
		}

		defaults, _ := config.LoadDefaults()
		if defaults[lang.Name] == "" {
			defaults[lang.Name] = resolvedVersion
			config.SaveDefaults(defaults)
			ui.Success("Set %s %s as default", lang.Name, resolvedVersion)
		}

		if err := install.CreateShims(); err != nil {
			ui.Warn("Warning: failed to update shims: %v", err)
		}
	},
}

func init() {
	installCmd.Flags().BoolP("verbose", "v", false, "Show build output")
}
