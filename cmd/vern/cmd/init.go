package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [language] [version]",
	Short: "Create a .vern file in the current directory",
	Long: `Create a .vern file for per-project version pinning.
If language and version are provided, writes them to the file.
Otherwise creates an empty .vern file for you to edit.`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		content := ""
		if len(args) >= 1 {
			lang := args[0]
			version := ""
			if len(args) == 2 {
				version = args[1]
			}
			if version != "" {
				content = fmt.Sprintf("%s %s\n", lang, version)
			} else {
				content = fmt.Sprintf("%s\n", lang)
			}
		}

		if err := os.WriteFile(".vern", []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating .vern file: %v\n", err)
			os.Exit(1)
		}

		if content == "" {
			fmt.Println("Created empty .vern file")
		} else {
			fmt.Printf("Created .vern file: %s", content)
		}
	},
}
