package internal

import (
	"github.com/spf13/cobra"
)

// CreateRootCmd creates the root command.
func CreateRootCmd() cobra.Command {
	rootCmd := cobra.Command{
		Use:   "gopuml",
		Short: "Compiles Plant UML files",
	}

	return rootCmd
}
