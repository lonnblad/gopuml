package internal

import (
	"fmt"

	"github.com/spf13/cobra"
)

// CreateVersionCmd creates the version subcommand.
func CreateVersionCmd(version string) cobra.Command {
	return cobra.Command{
		Use:   "version",
		Short: "Show current version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "gopuml version is:", version)
		},
		DisableFlagsInUseLine: true,
	}
}
