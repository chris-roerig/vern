package main

import (
	"os"

	"github.com/chris-roerig/vern/cmd/vern/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
