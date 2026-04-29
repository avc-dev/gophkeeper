package secret

import (
	"fmt"
	"os"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	"github.com/spf13/cobra"
)

func newAddBinaryCmd(app *cmdutil.App) *cobra.Command {
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

			masterKey, err := app.ResolveMasterKey(ctx, masterPwd)
			if err != nil {
				return err
			}
			defer cmdutil.ZeroKey(masterKey)

			authedCtx, err := app.AuthedContext(ctx)
			if err != nil {
				return fmt.Errorf("auth: %w", err)
			}

			if err := app.SecretSvc.AddBinary(authedCtx, masterKey, name, file, data, note); err != nil {
				return fmt.Errorf("add binary: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Binary %q saved (%d bytes).\n", name, len(data))
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name (required)")
	cmd.Flags().StringVar(&file, "file", "", "Path to file (required)")
	cmd.Flags().StringVar(&note, "note", "", "Note (optional)")
	cmdutil.AddMasterPasswordFlag(cmd, &masterPwd)
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}
