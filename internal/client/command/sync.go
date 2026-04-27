package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	var password string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Push pending changes and pull updates from the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if password == "" {
				var err error
				password, err = readPassword("Master password: ")
				if err != nil {
					return fmt.Errorf("read password: %w", err)
				}
			}

			masterKey, err := state.authService.DeriveMasterKey(ctx, password)
			if err != nil {
				return fmt.Errorf("derive master key: %w", err)
			}
			defer zeroKey(masterKey)

			authedCtx, err := authedContext(ctx)
			if err != nil {
				return fmt.Errorf("auth: %w", err)
			}

			if err := state.secretSvc.PushPending(authedCtx); err != nil {
				return fmt.Errorf("push failed: %w", err)
			}

			since, err := state.authService.GetLastSyncAt(ctx)
			if err != nil {
				return fmt.Errorf("read last_sync_at: %w", err)
			}

			if err := state.secretSvc.Sync(authedCtx, masterKey, since); err != nil {
				return fmt.Errorf("pull failed: %w", err)
			}

			if err := state.authService.SetLastSyncAt(ctx, nowUTC()); err != nil {
				return fmt.Errorf("update last_sync_at: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Sync completed.")
			return nil
		},
	}

	cmd.Flags().StringVar(&password, "master-password", "", "Master password (prompted if omitted)")
	return cmd
}
