package cmd

import (
	"fmt"

	"github.com/chris-roerig/vern/internal/ui"
	"github.com/spf13/cobra"
)

var Version = "1.6.0"

var rootCmd = &cobra.Command{
	Use:                "vern",
	Short:              "Vern - Version Number Manager",
	Long:               `Vern is a programming language version installation manager. It works like you think it should.`,
	CompletionOptions:  cobra.CompletionOptions{HiddenDefaultCmd: true},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.Logo)
		ui.Dim("v%s", Version)
		fmt.Println()
		cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(defaultCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(implodeCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(statsCmd)
}
