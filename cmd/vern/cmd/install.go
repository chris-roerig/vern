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

		// Check build dependencies from language config
		if lang.Requires != nil {
			vi, _ := version.ParseVersion(resolvedVersion)
			if lang.Requires.WhenMajorGte == 0 || vi.Major >= lang.Requires.WhenMajorGte {
				if !hasDependency(lang.Requires) {
					ui.Warn("%s %s requires %s %s+ to build.",
						lang.Name, resolvedVersion, lang.Requires.Binary, lang.Requires.MinVersion)
					fmt.Printf("Install %s %s now? [y/N]: ", lang.Name, lang.Requires.BootstrapVersion)
					scanner := bufio.NewScanner(os.Stdin)
					scanner.Scan()
					if strings.ToLower(scanner.Text()) == "y" {
						bootstrapVer, err := version.ResolveVersion(lang, lang.Requires.BootstrapVersion)
						if err != nil {
							ui.Error("Failed to resolve %s %s: %v", lang.Name, lang.Requires.BootstrapVersion, err)
							os.Exit(1)
						}
						ui.Info("Installing %s %s first...", lang.Name, bootstrapVer)
						if err := install.DownloadAndInstall(lang, bootstrapVer, opts); err != nil {
							ui.Error("Failed to install %s %s: %v", lang.Name, bootstrapVer, err)
							os.Exit(1)
						}
						ui.Success("%s %s installed.", lang.Name, bootstrapVer)
						binDir := filepath.Dir(filepath.Join(config.LanguageInstallDir(lang.Name, bootstrapVer), lang.Install.BinRelPath))
						os.Setenv("PATH", binDir+":"+os.Getenv("PATH"))
					} else {
						ui.Error("%s %s cannot be built without %s %s+",
							lang.Name, resolvedVersion, lang.Requires.Binary, lang.Requires.MinVersion)
						os.Exit(1)
					}
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

// hasDependency checks if a build requirement is satisfied by vern-installed
// or system binaries. Skips vern shims to avoid false negatives.
func hasDependency(req *config.Requirement) bool {
	// Check vern-installed versions first
	versions, _ := install.GetInstalledVersions(req.Binary)
	for _, v := range versions {
		if meetsMinVersion(v, req.MinVersion) {
			return true
		}
	}
	// Check system binary, skipping vern shims directory
	shimsDir := config.ShimsDir()
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == shimsDir {
			continue
		}
		bin := filepath.Join(dir, req.Binary)
		if _, err := os.Stat(bin); err != nil {
			continue
		}
		out, err := exec.Command(bin, "-e", "puts RUBY_VERSION").Output()
		if err != nil {
			out, err = exec.Command(bin, "--version").Output()
		}
		if err == nil {
			ver := extractVersion(strings.TrimSpace(string(out)))
			if ver != "" && meetsMinVersion(ver, req.MinVersion) {
				return true
			}
		}
	}
	return false
}

func meetsMinVersion(ver, minVer string) bool {
	v, err := version.ParseVersion(ver)
	if err != nil {
		return false
	}
	min, err := version.ParseVersion(minVer)
	if err != nil {
		return false
	}
	return v.Compare(min) >= 0
}

func extractVersion(s string) string {
	// Find first M.m.p pattern in output
	for _, word := range strings.Fields(s) {
		word = strings.TrimRight(word, ",;)")
		if config.IsValidVersion(word) && strings.Count(word, ".") == 2 {
			return word
		}
	}
	return ""
}
