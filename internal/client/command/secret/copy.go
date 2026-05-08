package secret

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	"github.com/spf13/cobra"
)

// NewCopyCmd возвращает cobra-команду "copy" для копирования поля секрета в буфер обмена.
func NewCopyCmd(app *cmdutil.App) *cobra.Command {
	var masterPwd, field string

	cmd := &cobra.Command{
		Use:   "copy <type> <name>",
		Short: "Copy a secret to the clipboard",
		Long: `Copies a sensitive field to the clipboard.
For credential: copies the password.
For card: copies the card number (--field number) or CVV (--field cvv).`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			typStr := args[0]
			name := args[1]

			masterKey, err := app.ResolveMasterKey(ctx, masterPwd)
			if err != nil {
				return err
			}
			defer cmdutil.ZeroKey(masterKey)

			var value string
			switch typStr {
			case "credential":
				p, err := app.SecretSvc.GetCredential(ctx, masterKey, name)
				if err != nil {
					return fmt.Errorf("get credential: %w", err)
				}
				value = p.Password
			case "card":
				p, err := app.SecretSvc.GetCard(ctx, masterKey, name)
				if err != nil {
					return fmt.Errorf("get card: %w", err)
				}
				switch field {
				case "cvv":
					value = p.CVV
				default:
					value = p.Number
				}
			default:
				return fmt.Errorf("copy is only supported for credential and card types")
			}

			if err := clipboard.WriteAll(value); err != nil {
				return fmt.Errorf("write to clipboard: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Copied to clipboard.")
			return nil
		},
	}

	cmdutil.AddMasterPasswordFlag(cmd, &masterPwd)
	cmd.Flags().StringVar(&field, "field", "password", "Field to copy for card: number, cvv")
	return cmd
}
