package command

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

func newCopyCmd() *cobra.Command {
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

			if masterPwd == "" {
				var err error
				masterPwd, err = readPassword("Master password: ")
				if err != nil {
					return fmt.Errorf("read password: %w", err)
				}
			}
			masterKey, err := state.authService.DeriveMasterKey(ctx, masterPwd)
			if err != nil {
				return fmt.Errorf("derive master key: %w", err)
			}
			defer zeroKey(masterKey)

			var value string
			switch typStr {
			case "credential":
				p, err := state.secretSvc.GetCredential(ctx, masterKey, name)
				if err != nil {
					return fmt.Errorf("get credential: %w", err)
				}
				value = p.Password
			case "card":
				p, err := state.secretSvc.GetCard(ctx, masterKey, name)
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

	cmd.Flags().StringVar(&masterPwd, "master-password", "", "Master password (prompted if omitted)")
	cmd.Flags().StringVar(&field, "field", "password", "Field to copy for card: number, cvv")
	return cmd
}
