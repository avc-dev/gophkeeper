package secret

import (
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	"github.com/spf13/cobra"
)

func newGetCredentialCmd(app *cmdutil.App) *cobra.Command {
	var masterPwd string

	cmd := &cobra.Command{
		Use:   "credential <name>",
		Short: "Show login and password",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

			masterKey, err := app.ResolveMasterKey(ctx, masterPwd)
			if err != nil {
				return err
			}
			defer cmdutil.ZeroKey(masterKey)

			p, err := app.SecretSvc.GetCredential(ctx, masterKey, name)
			if err != nil {
				return fmt.Errorf("get credential: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Name:     %s\n", name)
			fmt.Fprintf(out, "Login:    %s\n", p.Login)
			fmt.Fprintf(out, "Password: %s\n", p.Password)
			if p.URL != "" {
				fmt.Fprintf(out, "URL:      %s\n", p.URL)
			}
			if p.Note != "" {
				fmt.Fprintf(out, "Note:     %s\n", p.Note)
			}
			return nil
		},
	}
	cmdutil.AddMasterPasswordFlag(cmd, &masterPwd)
	return cmd
}
