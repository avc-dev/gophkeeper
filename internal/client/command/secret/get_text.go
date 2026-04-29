package secret

import (
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	"github.com/spf13/cobra"
)

func newGetTextCmd(app *cmdutil.App) *cobra.Command {
	var masterPwd string

	cmd := &cobra.Command{
		Use:   "text <name>",
		Short: "Show text note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

			masterKey, err := app.ResolveMasterKey(ctx, masterPwd)
			if err != nil {
				return err
			}
			defer cmdutil.ZeroKey(masterKey)

			p, err := app.SecretSvc.GetText(ctx, masterKey, name)
			if err != nil {
				return fmt.Errorf("get text: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), p.Content)
			return nil
		},
	}
	cmdutil.AddMasterPasswordFlag(cmd, &masterPwd)
	return cmd
}
