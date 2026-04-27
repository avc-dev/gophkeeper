package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in and sync secrets from the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if password == "" {
				var err error
				password, err = readPassword("Master password: ")
				if err != nil {
					return err
				}
			}

			masterKey, err := state.authService.Login(ctx, email, password)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}
			defer zeroKey(masterKey)

			// сразу синхронизируем все секреты с сервера.
			authedCtx, err := authedContext(ctx)
			if err != nil {
				return err
			}
			if err := state.secretSvc.Sync(authedCtx, masterKey, nil); err != nil {
				// не критично — локальный кеш мог быть пустым.
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

// zeroKey обнуляет master key в памяти после использования.
func zeroKey(key []byte) {
	for i := range key {
		key[i] = 0
	}
}
