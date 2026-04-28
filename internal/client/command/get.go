package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a secret by name",
	}
	cmd.AddCommand(
		newGetCredentialCmd(),
		newGetCardCmd(),
		newGetTextCmd(),
		newGetBinaryCmd(),
	)
	return cmd
}

func newGetCredentialCmd() *cobra.Command {
	var masterPwd string

	cmd := &cobra.Command{
		Use:   "credential <name>",
		Short: "Show login and password",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

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

			p, err := state.secretSvc.GetCredential(ctx, masterKey, name)
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
	cmd.Flags().StringVar(&masterPwd, "master-password", "", "Master password (prompted if omitted)")
	return cmd
}

func newGetCardCmd() *cobra.Command {
	var masterPwd string
	var showCVV bool

	cmd := &cobra.Command{
		Use:   "card <name>",
		Short: "Show card details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

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

			p, err := state.secretSvc.GetCard(ctx, masterKey, name)
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
	cmd.Flags().StringVar(&masterPwd, "master-password", "", "Master password (prompted if omitted)")
	cmd.Flags().BoolVar(&showCVV, "show-cvv", false, "Show CVV code")
	return cmd
}

func newGetTextCmd() *cobra.Command {
	var masterPwd string

	cmd := &cobra.Command{
		Use:   "text <name>",
		Short: "Show text note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

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

			p, err := state.secretSvc.GetText(ctx, masterKey, name)
			if err != nil {
				return fmt.Errorf("get text: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), p.Content)
			return nil
		},
	}
	cmd.Flags().StringVar(&masterPwd, "master-password", "", "Master password (prompted if omitted)")
	return cmd
}

func newGetBinaryCmd() *cobra.Command {
	var masterPwd, output string

	cmd := &cobra.Command{
		Use:   "binary <name>",
		Short: "Save binary file to disk",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

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

			p, err := state.secretSvc.GetBinary(ctx, masterKey, name)
			if err != nil {
				return fmt.Errorf("get binary: %w", err)
			}

			outPath := output
			if outPath == "" {
				outPath = p.Filename
			}
			if err := os.WriteFile(outPath, p.Data, 0o600); err != nil {
				return fmt.Errorf("write file: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Saved to %s (%d bytes).\n", outPath, len(p.Data))
			return nil
		},
	}
	cmd.Flags().StringVar(&masterPwd, "master-password", "", "Master password (prompted if omitted)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default: original filename)")
	return cmd
}

// maskCardNumber показывает только последние 4 цифры.
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
