package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd(version, buildTime string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version and build time",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "gophkeeper %s (built %s)\n", version, buildTime)
			return nil
		},
	}
}
