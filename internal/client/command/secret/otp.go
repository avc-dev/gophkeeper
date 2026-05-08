package secret

import (
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	svc "github.com/avc-dev/gophkeeper/internal/client/service/secret"
	"github.com/spf13/cobra"
)

// NewOTPCmd возвращает cobra-команду "otp" — генерирует текущий TOTP-код для сохранённого семени.
func NewOTPCmd(app *cmdutil.App) *cobra.Command {
	var masterPwd string

	cmd := &cobra.Command{
		Use:   "otp <name>",
		Short: "Generate a TOTP code for a saved OTP secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

			masterKey, err := app.ResolveMasterKey(ctx, masterPwd)
			if err != nil {
				return err
			}
			defer cmdutil.ZeroKey(masterKey)

			payload, err := app.SecretSvc.GetOTP(ctx, masterKey, name)
			if err != nil {
				return fmt.Errorf("get otp: %w", err)
			}

			code, secondsLeft, err := svc.GenerateOTP(payload.Seed)
			if err != nil {
				return fmt.Errorf("generate code: %w", err)
			}

			out := cmd.OutOrStdout()
			// Форматируем код как "123 456" для удобства чтения
			if len(code) == 6 {
				fmt.Fprintf(out, "Code:    %s %s\n", code[:3], code[3:])
			} else {
				fmt.Fprintf(out, "Code:    %s\n", code)
			}
			fmt.Fprintf(out, "Expires: %ds\n", secondsLeft)
			if payload.Issuer != "" || payload.Account != "" {
				fmt.Fprintf(out, "Account: %s", payload.Issuer)
				if payload.Account != "" {
					if payload.Issuer != "" {
						fmt.Fprintf(out, " (%s)", payload.Account)
					} else {
						fmt.Fprint(out, payload.Account)
					}
				}
				fmt.Fprintln(out)
			}
			return nil
		},
	}

	cmdutil.AddMasterPasswordFlag(cmd, &masterPwd)
	return cmd
}
