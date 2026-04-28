package auth

import (
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	"github.com/spf13/cobra"
)

func NewRegisterCmd(app *cmdutil.App) *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if password == "" {
				var err error
				password, err = cmdutil.ReadPassword("Master password: ")
				if err != nil {
					return fmt.Errorf("read password: %w", err)
				}
			}

			if err := app.AuthSvc.Register(ctx, email, password); err != nil {
				return fmt.Errorf("registration failed: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Registration successful. Run 'gophkeeper login' to continue.")
			fmt.Fprintln(cmd.OutOrStdout(), "⚠  Remember your master password — it cannot be recovered.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Master password (prompted if omitted)")
	_ = cmd.MarkFlagRequired("email")

	return cmd
}
