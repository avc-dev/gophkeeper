package secret

import (
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	"github.com/spf13/cobra"
)

func NewSyncCmd(app *cmdutil.App) *cobra.Command {
	var masterPwd string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Push pending changes and pull updates from the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			masterKey, err := app.ResolveMasterKey(ctx, masterPwd)
			if err != nil {
				return err
			}
			defer cmdutil.ZeroKey(masterKey)

			authedCtx, err := app.AuthedContext(ctx)
			if err != nil {
				return fmt.Errorf("auth: %w", err)
			}

			if err := app.SecretSvc.PushPending(authedCtx); err != nil {
				return fmt.Errorf("push failed: %w", err)
			}

			since, err := app.AuthSvc.GetLastSyncAt(ctx)
			if err != nil {
				return fmt.Errorf("read last_sync_at: %w", err)
			}

			if err := app.SecretSvc.Sync(authedCtx, masterKey, since); err != nil {
				return fmt.Errorf("pull failed: %w", err)
			}

			if err := app.AuthSvc.SetLastSyncAt(ctx, cmdutil.NowUTC()); err != nil {
				return fmt.Errorf("update last_sync_at: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Sync completed.")
			return nil
		},
	}

	cmdutil.AddMasterPasswordFlag(cmd, &masterPwd)
	return cmd
}
