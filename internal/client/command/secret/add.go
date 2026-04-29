package secret

import (
	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	"github.com/spf13/cobra"
)

// NewAddCmd возвращает cobra-команду "add" с подкомандами credential/card/text/binary.
func NewAddCmd(app *cmdutil.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a secret",
	}
	cmd.AddCommand(
		newAddCredentialCmd(app),
		newAddCardCmd(app),
		newAddTextCmd(app),
		newAddBinaryCmd(app),
	)
	return cmd
}
