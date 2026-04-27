package command

import (
	"fmt"
	"os"

	"github.com/avc-dev/gophkeeper/internal/client/service"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a secret",
	}
	cmd.AddCommand(
		newAddCredentialCmd(),
		newAddCardCmd(),
		newAddTextCmd(),
		newAddBinaryCmd(),
	)
	return cmd
}

func newAddCredentialCmd() *cobra.Command {
	var name, login, password, url, note, masterPwd string

	cmd := &cobra.Command{
		Use:   "credential",
		Short: "Add a login/password credential",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

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

			authedCtx, err := authedContext(ctx)
			if err != nil {
				return fmt.Errorf("auth: %w", err)
			}

			if err := state.secretSvc.AddCredential(authedCtx, masterKey, name, login, password, url, note); err != nil {
				return fmt.Errorf("add credential: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Credential %q saved.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Unique name (required)")
	cmd.Flags().StringVarP(&login, "login", "l", "", "Login / username (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password (required)")
	cmd.Flags().StringVarP(&url, "url", "u", "", "URL (optional)")
	cmd.Flags().StringVar(&note, "note", "", "Note (optional)")
	cmd.Flags().StringVar(&masterPwd, "master-password", "", "Master password (prompted if omitted)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("login")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}

func newAddCardCmd() *cobra.Command {
	var name, number, holder, expiry, cvv, bank, note, masterPwd string

	cmd := &cobra.Command{
		Use:   "card",
		Short: "Add a bank card",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := service.ValidateCard(number, expiry, cvv); err != nil {
				return err // sentinel errors с понятным сообщением
			}

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

			authedCtx, err := authedContext(ctx)
			if err != nil {
				return fmt.Errorf("auth: %w", err)
			}

			if err := state.secretSvc.AddCard(authedCtx, masterKey, name, number, holder, expiry, cvv, bank, note); err != nil {
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
	cmd.Flags().StringVar(&masterPwd, "master-password", "", "Master password (prompted if omitted)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("number")
	_ = cmd.MarkFlagRequired("holder")
	_ = cmd.MarkFlagRequired("expiry")
	_ = cmd.MarkFlagRequired("cvv")
	return cmd
}

func newAddTextCmd() *cobra.Command {
	var name, content, file, note, masterPwd string

	cmd := &cobra.Command{
		Use:   "text",
		Short: "Add a text note",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if content == "" && file == "" {
				return fmt.Errorf("provide --content or --file")
			}
			if file != "" {
				data, err := os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("read file: %w", err)
				}
				content = string(data)
			}

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

			authedCtx, err := authedContext(ctx)
			if err != nil {
				return fmt.Errorf("auth: %w", err)
			}

			if err := state.secretSvc.AddText(authedCtx, masterKey, name, content, note); err != nil {
				return fmt.Errorf("add text: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Text %q saved.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name (required)")
	cmd.Flags().StringVar(&content, "content", "", "Text content")
	cmd.Flags().StringVar(&file, "file", "", "Read content from file")
	cmd.Flags().StringVar(&note, "note", "", "Note (optional)")
	cmd.Flags().StringVar(&masterPwd, "master-password", "", "Master password (prompted if omitted)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newAddBinaryCmd() *cobra.Command {
	var name, file, note, masterPwd string

	cmd := &cobra.Command{
		Use:   "binary",
		Short: "Add a binary file",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			if masterPwd == "" {
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

			authedCtx, err := authedContext(ctx)
			if err != nil {
				return fmt.Errorf("auth: %w", err)
			}

			if err := state.secretSvc.AddBinary(authedCtx, masterKey, name, file, data, note); err != nil {
				return fmt.Errorf("add binary: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Binary %q saved (%d bytes).\n", name, len(data))
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name (required)")
	cmd.Flags().StringVar(&file, "file", "", "Path to file (required)")
	cmd.Flags().StringVar(&note, "note", "", "Note (optional)")
	cmd.Flags().StringVar(&masterPwd, "master-password", "", "Master password (prompted if omitted)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}
