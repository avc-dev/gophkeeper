package secret

import (
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	secretsvc "github.com/avc-dev/gophkeeper/internal/client/service/secret"
	"github.com/spf13/cobra"
)

func newAddCardCmd(app *cmdutil.App) *cobra.Command {
	var name, number, holder, expiry, cvv, bank, note, masterPwd string

	cmd := &cobra.Command{
		Use:   "card",
		Short: "Add a bank card",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := secretsvc.ValidateCard(number, expiry, cvv); err != nil {
				return err
			}

			masterKey, err := app.ResolveMasterKey(ctx, masterPwd)
			if err != nil {
				return err
			}
			defer cmdutil.ZeroKey(masterKey)

			authedCtx, err := app.AuthedContext(ctx)
			if err != nil {
				return fmt.Errorf("auth: %w", err)
			}

			if err := app.SecretSvc.AddCard(authedCtx, masterKey, name, number, holder, expiry, cvv, bank, note); err != nil {
				return fmt.Errorf("add card: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Card %q saved.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Card name (required)")
	cmd.Flags().StringVar(&number, "number", "", "Card number (required)")
	cmd.Flags().StringVar(&holder, "holder", "", "Card holder name (required)")
	cmd.Flags().StringVar(&expiry, "expiry", "", "Expiry date MM/YY (required)")
	cmd.Flags().StringVar(&cvv, "cvv", "", "CVV/CVC (required)")
	cmd.Flags().StringVar(&bank, "bank", "", "Bank name (optional)")
	cmd.Flags().StringVar(&note, "note", "", "Note (optional)")
	cmdutil.AddMasterPasswordFlag(cmd, &masterPwd)
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("number")
	_ = cmd.MarkFlagRequired("holder")
	_ = cmd.MarkFlagRequired("expiry")
	_ = cmd.MarkFlagRequired("cvv")
	return cmd
}
