// Package command содержит cobra-команды CLI-клиента GophKeeper.
package command

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"os"

	authcmd "github.com/avc-dev/gophkeeper/internal/client/command/auth"
	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	secretcmd "github.com/avc-dev/gophkeeper/internal/client/command/secret"
	"github.com/avc-dev/gophkeeper/internal/client/config"
	authsvc "github.com/avc-dev/gophkeeper/internal/client/service/auth"
	secretsvc "github.com/avc-dev/gophkeeper/internal/client/service/secret"
	"github.com/avc-dev/gophkeeper/internal/client/storage"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// NewRootCmd строит и возвращает корневую cobra-команду.
func NewRootCmd(version, buildTime string) *cobra.Command {
	app := &cmdutil.App{}

	var (
		db   *sql.DB
		conn *grpc.ClientConn
	)

	root := &cobra.Command{
		Use:   "gophkeeper",
		Short: "Zero-knowledge password manager",
		Long: `GophKeeper — client-server password manager.
All secrets are encrypted on the client before being sent to the server.`,
		SilenceUsage: true,
	}

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "version" || cmd.Name() == "help" {
			return nil
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		db, err = storage.Open(cfg.DBPath)
		if err != nil {
			return fmt.Errorf("open local db: %w", err)
		}

		dialCreds, err := buildTransportCredentials(cfg)
		if err != nil {
			_ = db.Close()
			return fmt.Errorf("tls credentials: %w", err)
		}

		conn, err = grpc.NewClient(cfg.ServerAddr, grpc.WithTransportCredentials(dialCreds))
		if err != nil {
			_ = db.Close()
			return fmt.Errorf("connect to server %s: %w", cfg.ServerAddr, err)
		}

		app.AuthSvc = authsvc.New(pb.NewAuthServiceClient(conn), storage.NewAuthStorage(db))
		app.SecretSvc = secretsvc.New(pb.NewSecretsServiceClient(conn), storage.NewSecretStorage(db))
		return nil
	}

	root.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		if db == nil {
			return nil
		}
		_ = storage.Checkpoint(cmd.Context(), db)
		_ = conn.Close()
		_ = db.Close()
		db = nil
		return nil
	}

	root.AddCommand(
		authcmd.NewRegisterCmd(app),
		authcmd.NewLoginCmd(app),
		secretcmd.NewSyncCmd(app),
		secretcmd.NewAddCmd(app),
		secretcmd.NewGetCmd(app),
		secretcmd.NewListCmd(app),
		secretcmd.NewDeleteCmd(app),
		secretcmd.NewCopyCmd(app),
		newVersionCmd(version, buildTime),
	)
	return root
}

// buildTransportCredentials возвращает транспортные credentials на основе конфигурации TLS.
func buildTransportCredentials(cfg config.Config) (credentials.TransportCredentials, error) {
	if !cfg.TLSEnabled {
		return insecure.NewCredentials(), nil
	}

	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}

	if cfg.TLSCACert != "" {
		caPEM, err := os.ReadFile(cfg.TLSCACert)
		if err != nil {
			return nil, fmt.Errorf("read CA cert %q: %w", cfg.TLSCACert, err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("invalid CA cert %q: no PEM blocks found", cfg.TLSCACert)
		}
		tlsCfg.RootCAs = pool
	}

	if cfg.TLSSkipVerify {
		tlsCfg.InsecureSkipVerify = true //nolint:gosec // допустимо только для разработки с self-signed cert
	}

	return credentials.NewTLS(tlsCfg), nil
}
