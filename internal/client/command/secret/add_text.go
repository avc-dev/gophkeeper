package secret

import (
	"fmt"
	"os"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	"github.com/spf13/cobra"
)

func newAddTextCmd(app *cmdutil.App) *cobra.Command {
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

			masterKey, err := app.ResolveMasterKey(ctx, masterPwd)
			if err != nil {
				return err
			}
			defer cmdutil.ZeroKey(masterKey)

			authedCtx, err := app.AuthedContext(ctx)
			if err != nil {
				return fmt.Errorf("auth: %w", err)
			}

			if err := app.SecretSvc.AddText(authedCtx, masterKey, name, content, note); err != nil {
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
	cmdutil.AddMasterPasswordFlag(cmd, &masterPwd)
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
