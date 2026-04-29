package secret

import (
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	svc "github.com/avc-dev/gophkeeper/internal/client/service/secret"
	"github.com/spf13/cobra"
)

func newAddOTPCmd(app *cmdutil.App) *cobra.Command {
	var name, seed, issuer, account, note, masterPwd string

	cmd := &cobra.Command{
		Use:   "otp",
		Short: "Add a TOTP secret (2FA seed)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := svc.ValidateOTPSeed(seed); err != nil {
				return fmt.Errorf("invalid seed: %w", err)
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

			if err := app.SecretSvc.AddOTP(authedCtx, masterKey, name, seed, issuer, account, note); err != nil {
				return fmt.Errorf("add otp: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "OTP %q saved.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Unique name (required)")
	cmd.Flags().StringVarP(&seed, "seed", "s", "", "Base32-encoded TOTP secret key (required)")
	cmd.Flags().StringVar(&issuer, "issuer", "", "Service name, e.g. GitHub (optional)")
	cmd.Flags().StringVar(&account, "account", "", "Account / email (optional)")
	cmd.Flags().StringVar(&note, "note", "", "Note (optional)")
	cmdutil.AddMasterPasswordFlag(cmd, &masterPwd)
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("seed")
	return cmd
}
