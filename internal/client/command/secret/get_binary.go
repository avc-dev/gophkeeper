package secret

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	"github.com/spf13/cobra"
)

func newGetBinaryCmd(app *cmdutil.App) *cobra.Command {
	var masterPwd, output string

	cmd := &cobra.Command{
		Use:   "binary <name>",
		Short: "Save binary file to disk",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

			masterKey, err := app.ResolveMasterKey(ctx, masterPwd)
			if err != nil {
				return err
			}
			defer cmdutil.ZeroKey(masterKey)

			p, err := app.SecretSvc.GetBinary(ctx, masterKey, name)
			if err != nil {
				return fmt.Errorf("get binary: %w", err)
			}

			rawData, err := base64.StdEncoding.DecodeString(p.Data)
			if err != nil {
				return fmt.Errorf("decode binary data: %w", err)
			}

			outPath := output
			if outPath == "" {
				outPath = p.Filename
			}
			if err := os.WriteFile(outPath, rawData, 0o600); err != nil {
				return fmt.Errorf("write file: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Saved to %s (%d bytes).\n", outPath, len(rawData))
			return nil
		},
	}
	cmdutil.AddMasterPasswordFlag(cmd, &masterPwd)
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default: original filename)")
	return cmd
}
