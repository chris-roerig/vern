package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "0.7.0"

var rootCmd = &cobra.Command{
	Use:   "vern",
	Short: "Vern - Version Number Manager",
	Long:  `Vern is a programming language version installation manager that just makes sense.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("vern %s\n\n", Version)
		fmt.Println(cmd.Help())
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
	rootCmd.AddCommand(helpCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(implodeCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(statsCmd)
}
