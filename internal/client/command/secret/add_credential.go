package secret

import (
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	"github.com/spf13/cobra"
)

func newAddCredentialCmd(app *cmdutil.App) *cobra.Command {
	var name, login, password, url, note, masterPwd string

	cmd := &cobra.Command{
		Use:   "credential",
		Short: "Add a login/password credential",
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

			if err := app.SecretSvc.AddCredential(authedCtx, masterKey, name, login, password, url, note); err != nil {
				return fmt.Errorf("add credential: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Credential %q saved.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Unique name (required)")
	cmd.Flags().StringVarP(&login, "login", "l", "", "Login / username (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password (required)")
	cmd.Flags().StringVarP(&url, "url", "u", "", "URL (optional)")
	cmd.Flags().StringVar(&note, "note", "", "Note (optional)")
	cmdutil.AddMasterPasswordFlag(cmd, &masterPwd)
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("login")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}
