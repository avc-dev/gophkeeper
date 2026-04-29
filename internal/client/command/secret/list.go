package secret

import (
	"fmt"
	"text/tabwriter"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/spf13/cobra"
)

func NewListCmd(app *cmdutil.App) *cobra.Command {
	var typFilter string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List secrets from local cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var typ domain.SecretType
			if typFilter != "" {
				typ = domain.SecretType(typFilter)
			}

			secrets, err := app.SecretSvc.List(ctx, typ)
			if err != nil {
				return fmt.Errorf("list secrets: %w", err)
			}

			if len(secrets) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No secrets found.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tTYPE\tSTATUS\tUPDATED")
			for _, s := range secrets {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					s.Name,
					string(s.Type),
					string(s.SyncStatus),
					s.UpdatedAt.Format("2006-01-02 15:04"),
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVarP(&typFilter, "type", "t", "",
		"Filter by type: credential, card, text, binary")
	return cmd
}
