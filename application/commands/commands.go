package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCommand = &cobra.Command{
		Use:   "admirer",
		Short: "A command line utility to sync loved tracks between music services.",
	}
	limit int
	page  int
)

// Execute runs the requested CLI command.
func Execute() {
	err := rootCommand.Execute()
	if err != nil {
		os.Exit(1)
	}
}
