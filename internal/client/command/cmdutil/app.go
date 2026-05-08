package cmdutil

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/avc-dev/gophkeeper/internal/client/service"
	authsvc "github.com/avc-dev/gophkeeper/internal/client/service/auth"
	secretsvc "github.com/avc-dev/gophkeeper/internal/client/service/secret"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// App holds initialized service dependencies.
// Allocated empty before cobra setup; filled by PersistentPreRunE at runtime.
type App struct {
	AuthSvc   *authsvc.Service
	SecretSvc *secretsvc.Service
}

// AuthedContext builds a context carrying the user's JWT for outgoing gRPC calls.
func (a *App) AuthedContext(ctx context.Context) (context.Context, error) {
	token, err := a.AuthSvc.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("build auth context: %w", err)
	}
	return service.ContextWithBearerToken(ctx, token), nil
}

// ResolveMasterKey prompts for the master password if pwd is empty, then derives the encryption key.
func (a *App) ResolveMasterKey(ctx context.Context, pwd string) ([]byte, error) {
	if pwd == "" {
		var err error
		pwd, err = ReadPassword("Master password: ")
		if err != nil {
			return nil, fmt.Errorf("read password: %w", err)
		}
	}
	key, err := a.AuthSvc.DeriveMasterKey(ctx, pwd)
	if err != nil {
		return nil, fmt.Errorf("derive master key: %w", err)
	}
	return key, nil
}

// ReadPassword reads a password from the terminal without echo.
// Falls back to plain stdin read when not in a terminal (pipes, tests).
func ReadPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", fmt.Errorf("read password: %w", err)
		}
		return string(b), nil
	}
	var s string
	_, err := fmt.Fscan(os.Stdin, &s)
	return s, err
}

// ZeroKey zeroes a master key in memory after use.
func ZeroKey(key []byte) { service.ZeroKey(key) }

// AddMasterPasswordFlag registers the --master-password flag on a command.
func AddMasterPasswordFlag(cmd *cobra.Command, target *string) {
	cmd.Flags().StringVar(target, "master-password", "", "Master password (prompted if omitted)")
}

// NowUTC returns the current UTC time.
func NowUTC() time.Time { return time.Now().UTC() }
