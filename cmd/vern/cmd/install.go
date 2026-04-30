package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chris-roerig/vern/internal/config"
	"github.com/chris-roerig/vern/internal/install"
	"github.com/chris-roerig/vern/internal/ui"
	"github.com/chris-roerig/vern/internal/version"
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
		opts := install.Options{Verbose: verbose}

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

		// Ruby 4.x requires Ruby 3.1+ to bootstrap the build
		if lang.Name == "ruby" {
			vi, _ := version.ParseVersion(resolvedVersion)
			if vi.Major >= 4 && !hasRuby31() {
				ui.Warn("Ruby %s requires Ruby 3.1+ to build.", resolvedVersion)
				fmt.Print("Install Ruby 3.4 now? [y/N]: ")
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				if strings.ToLower(scanner.Text()) == "y" {
					bootstrapVer, err := version.ResolveVersion(lang, "3.4")
					if err != nil {
						ui.Error("Failed to resolve Ruby 3.4: %v", err)
						os.Exit(1)
					}
					ui.Info("Installing Ruby %s first...", bootstrapVer)
					if err := install.DownloadAndInstall(lang, bootstrapVer, opts); err != nil {
						ui.Error("Failed to install Ruby %s: %v", bootstrapVer, err)
						os.Exit(1)
					}
					ui.Success("Ruby %s installed.", bootstrapVer)
					// Update PATH so the build can find it
					binDir := filepath.Dir(filepath.Join(config.LanguageInstallDir(lang.Name, bootstrapVer), lang.Install.BinRelPath))
					os.Setenv("PATH", binDir+":"+os.Getenv("PATH"))
				} else {
					ui.Error("Ruby 4.x cannot be built without Ruby 3.1+")
					os.Exit(1)
				}
			}
		}

		if err := install.DownloadAndInstall(lang, resolvedVersion, opts); err != nil {
			ui.Error("Error: %v", err)
			os.Exit(1)
		}

		defaults, _ := config.LoadDefaults()
		if defaults[lang.Name] == "" {
			defaults[lang.Name] = resolvedVersion
			if err := config.SaveDefaults(defaults); err != nil {
				ui.Warn("Failed to save defaults: %v", err)
			}
			ui.Success("Set %s %s as default", lang.Name, resolvedVersion)
		}

		if err := install.CreateShims(); err != nil {
			ui.Warn("Failed to update shims: %v", err)
		}
	},
}

func init() {
	installCmd.Flags().BoolP("verbose", "v", false, "Show build output")
}

// hasRuby31 checks if Ruby 3.1+ is available on PATH or installed via vern.
func hasRuby31() bool {
	// Check system ruby
	out, err := exec.Command("ruby", "-e", "puts RUBY_VERSION").Output()
	if err == nil {
		ver := strings.TrimSpace(string(out))
		if vi, err := version.ParseVersion(ver); err == nil {
			if vi.Major > 3 || (vi.Major == 3 && vi.Minor >= 1) {
				return true
			}
		}
	}
	// Check vern-installed ruby
	versions, _ := install.GetInstalledVersions("ruby")
	for _, v := range versions {
		if vi, err := version.ParseVersion(v); err == nil {
			if vi.Major > 3 || (vi.Major == 3 && vi.Minor >= 1) {
				return true
			}
		}
	}
	return false
}
