package command

import (
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <type> <name>",
		Short: "Delete a secret",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			typ := domain.SecretType(args[0])
			name := args[1]

			authedCtx, err := authedContext(ctx)
			if err != nil {
				return err
			}

			if err := state.secretSvc.Delete(authedCtx, name, typ); err != nil {
				return fmt.Errorf("delete secret: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Secret %q deleted.\n", name)
			return nil
		},
	}
	return cmd
}
