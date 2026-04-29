package cmd

import (
	"fmt"
	"os"

	"github.com/chris/vern/internal/install"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up shims for version switching",
	Long:  `Create shim scripts in the vern data directory for all supported languages.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := install.CreateShims(); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating shims: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Shims created successfully.")

		if !install.IsPathSet() {
			fmt.Println("\nAdd shims to your PATH:")
			fmt.Println(install.GetShellHook())
		}
	},
}
