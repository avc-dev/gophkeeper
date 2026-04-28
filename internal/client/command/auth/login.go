package auth

import (
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	"github.com/spf13/cobra"
)

func NewLoginCmd(app *cmdutil.App) *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in and sync secrets from the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if password == "" {
				var err error
				password, err = cmdutil.ReadPassword("Master password: ")
				if err != nil {
					return fmt.Errorf("read password: %w", err)
				}
			}

			masterKey, err := app.AuthSvc.Login(ctx, email, password)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}
			defer cmdutil.ZeroKey(masterKey)

			authedCtx, err := app.AuthedContext(ctx)
			if err != nil {
				return fmt.Errorf("auth: %w", err)
			}
			if err := app.SecretSvc.Sync(authedCtx, masterKey, nil); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: sync failed: %v\n", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Logged in successfully.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Master password (prompted if omitted)")
	_ = cmd.MarkFlagRequired("email")

	return cmd
}
