// Package command содержит cobra-команды CLI-клиента GophKeeper.
package command

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/avc-dev/gophkeeper/internal/client/config"
	"github.com/avc-dev/gophkeeper/internal/client/service"
	"github.com/avc-dev/gophkeeper/internal/client/storage"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// app — состояние приложения, инициализируется в PersistentPreRunE.
type app struct {
	db          *sql.DB
	authService *service.AuthService
	secretSvc   *service.SecretService
	conn        *grpc.ClientConn
	cfg         config.Config
}

var state *app

// NewRootCmd строит и возвращает корневую cobra-команду.
func NewRootCmd(version, buildTime string) *cobra.Command {
	root := &cobra.Command{
		Use:   "gophkeeper",
		Short: "Zero-knowledge password manager",
		Long: `GophKeeper — client-server password manager.
All secrets are encrypted on the client before being sent to the server.`,
		SilenceUsage: true,
	}

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// команды, не требующие инициализации (version, help).
		if cmd.Name() == "version" || cmd.Name() == "help" {
			return nil
		}
		return initApp(cmd.Context())
	}
	root.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		return cleanupApp(cmd.Context())
	}

	root.AddCommand(
		newRegisterCmd(),
		newLoginCmd(),
		newAddCmd(),
		newGetCmd(),
		newListCmd(),
		newDeleteCmd(),
		newCopyCmd(),
		newVersionCmd(version, buildTime),
	)
	return root
}

// initApp открывает БД и gRPC соединение.
func initApp(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := storage.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open local db: %w", err)
	}

	conn, err := grpc.NewClient(cfg.ServerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		db.Close()
		return fmt.Errorf("connect to server %s: %w", cfg.ServerAddr, err)
	}

	authStore := storage.NewAuthStorage(db)
	secretStore := storage.NewSecretStorage(db)

	authSvc := service.NewAuthService(pb.NewAuthServiceClient(conn), authStore)
	secretSvc := service.NewSecretService(pb.NewSecretsServiceClient(conn), secretStore)

	state = &app{
		db:          db,
		conn:        conn,
		cfg:         cfg,
		authService: authSvc,
		secretSvc:   secretSvc,
	}
	return nil
}

// cleanupApp выполняет WAL checkpoint и закрывает соединения.
func cleanupApp(ctx context.Context) error {
	if state == nil {
		return nil
	}
	_ = storage.Checkpoint(ctx, state.db)
	_ = state.conn.Close()
	_ = state.db.Close()
	state = nil
	return nil
}

// authedContext добавляет JWT токен в gRPC метаданные.
func authedContext(ctx context.Context) (context.Context, error) {
	token, err := state.authService.Token(ctx)
	if err != nil {
		return nil, err
	}
	md := metadata.Pairs("authorization", "Bearer "+token)
	return metadata.NewOutgoingContext(ctx, md), nil
}

// readPassword читает пароль из терминала без эха.
// Если stdin не терминал (pipe/тест) — читает как обычный ввод.
func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", fmt.Errorf("read password: %w", err)
		}
		return string(b), nil
	}
	// в не-терминальном режиме (тесты, pipe) читаем строку обычно.
	var s string
	_, err := fmt.Fscan(os.Stdin, &s)
	return s, err
}
