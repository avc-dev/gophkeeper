package secret

import (
	"fmt"
	"strings"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	"github.com/spf13/cobra"
)

func newGetCardCmd(app *cmdutil.App) *cobra.Command {
	var masterPwd string
	var showCVV bool

	cmd := &cobra.Command{
		Use:   "card <name>",
		Short: "Show card details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

			masterKey, err := app.ResolveMasterKey(ctx, masterPwd)
			if err != nil {
				return err
			}
			defer cmdutil.ZeroKey(masterKey)

			p, err := app.SecretSvc.GetCard(ctx, masterKey, name)
			if err != nil {
				return fmt.Errorf("get card: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Name:   %s\n", name)
			fmt.Fprintf(out, "Number: %s\n", maskCardNumber(p.Number, showCVV))
			fmt.Fprintf(out, "Holder: %s\n", p.Holder)
			fmt.Fprintf(out, "Expiry: %s\n", p.Expiry)
			if showCVV {
				fmt.Fprintf(out, "CVV:    %s\n", p.CVV)
			} else {
				fmt.Fprintf(out, "CVV:    ***\n")
			}
			if p.Bank != "" {
				fmt.Fprintf(out, "Bank:   %s\n", p.Bank)
			}
			if p.Note != "" {
				fmt.Fprintf(out, "Note:   %s\n", p.Note)
			}
			return nil
		},
	}
	cmdutil.AddMasterPasswordFlag(cmd, &masterPwd)
	cmd.Flags().BoolVar(&showCVV, "show-cvv", false, "Show CVV code")
	return cmd
}

func maskCardNumber(number string, full bool) string {
	if full {
		return number
	}
	clean := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, number)
	if len(clean) <= 4 {
		return number
	}
	return "**** **** **** " + clean[len(clean)-4:]
}
