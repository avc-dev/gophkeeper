package secret

import (
	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	"github.com/spf13/cobra"
)

// NewGetCmd возвращает cobra-команду "get" с подкомандами credential/card/text/binary.
func NewGetCmd(app *cmdutil.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a secret by name",
	}
	cmd.AddCommand(
		newGetCredentialCmd(app),
		newGetCardCmd(app),
		newGetTextCmd(app),
		newGetBinaryCmd(app),
	)
	return cmd
}
