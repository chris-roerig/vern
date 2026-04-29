package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Print help message",
	Long: `Print detailed help for vern commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(`vern - Version Number Manager (v0.1.0)

A programming language version installation manager that just makes sense.

USAGE:
    vern <command> [arguments]

COMMANDS:
    install <language> [version]   Install a language version
                                   (no version = latest, partial = latest match)
    list [language]                List installed versions
    remove <language>              Interactively remove installed versions
    default <language> [version]   Set or show default version for a language
    help                           Print this help message

EXAMPLES:
    vern install python                    # Install latest Python
    vern install python 3                  # Install latest Python 3.x.x
    vern install python 3.11               # Install latest Python 3.11.x
    vern install python 3.11.2             # Install specific version
    vern list                              # List all installed languages
    vern list python                       # List installed Python versions
    vern default python                    # Select default Python version
    vern default python 3.11.2             # Set Python 3.11.2 as default
    vern remove python                      # Interactively remove Python versions

VERSION SWITCHING:
    Create a .vern file in your project root with "language version"
    Example: echo "python 3.11.2" > .vern

    Vern resolves versions in this order:
    1. .vern file in current or parent directory
    2. Global default for the language

SHELL SETUP:
    Add vern shims to your PATH to enable version switching:
    export PATH="$HOME/.local/share/vern/shims:$PATH"

    Or run: vern setup

`)
		os.Exit(0)
	},
}
