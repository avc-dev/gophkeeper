package app

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/avc-dev/gophkeeper/internal/server/config"
	"github.com/avc-dev/gophkeeper/internal/server/handler"
	authhandler "github.com/avc-dev/gophkeeper/internal/server/handler/auth"
	secrethandler "github.com/avc-dev/gophkeeper/internal/server/handler/secret"
	authsvc "github.com/avc-dev/gophkeeper/internal/server/service/auth"
	secretsvc "github.com/avc-dev/gophkeeper/internal/server/service/secret"
	secretstore "github.com/avc-dev/gophkeeper/internal/server/storage/secret"
	userstore "github.com/avc-dev/gophkeeper/internal/server/storage/user"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Run собирает зависимости и запускает gRPC сервер.
// Блокируется до получения SIGINT/SIGTERM, после чего выполняет graceful stop.
func Run(cfg config.Config, log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := initDB(ctx, cfg.DSN)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}
	defer db.Close()

	privKey, pubKey, err := loadEdKeys(cfg.JWTPrivateKeyFile, cfg.JWTPublicKeyFile)
	if err != nil {
		return fmt.Errorf("load JWT keys: %w", err)
	}

	users := userstore.New(db)
	auth := authsvc.New(users, privKey, pubKey)

	secrets := secretstore.New(db)
	secretService := secretsvc.New(secrets)

	serverOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(handler.AuthInterceptor(auth)),
	}

	if cfg.TLSCertFile != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			return fmt.Errorf("load TLS credentials: %w", err)
		}
		serverOpts = append(serverOpts, grpc.Creds(creds))
		log.Info("TLS enabled", "cert", cfg.TLSCertFile)
	} else {
		serverOpts = append(serverOpts, grpc.Creds(insecure.NewCredentials()))
		log.Warn("TLS is disabled — server running in insecure mode")
	}

	srv := grpc.NewServer(serverOpts...)
	pb.RegisterAuthServiceServer(srv, authhandler.New(auth))
	pb.RegisterSecretsServiceServer(srv, secrethandler.New(secretService))

	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.Addr, err)
	}

	serveErr := make(chan error, 1)
	go func() {
		if err := srv.Serve(lis); err != nil {
			serveErr <- fmt.Errorf("grpc serve: %w", err)
		}
	}()

	log.Info("server is listening", "addr", cfg.Addr)

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		log.Info("shutting down gracefully")
		srv.GracefulStop()
		return nil
	}
}

// loadEdKeys читает Ed25519 ключевую пару из PEM-файлов.
// Приватный ключ ожидается в формате "PRIVATE KEY" (PKCS#8).
// Публичный ключ ожидается в формате "PUBLIC KEY" (PKIX/SubjectPublicKeyInfo).
// Генерация: openssl genpkey -algorithm ed25519 -out private.pem
//
//	openssl pkey -in private.pem -pubout -out public.pem
func loadEdKeys(privPath, pubPath string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	privPEM, err := os.ReadFile(privPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read private key %q: %w", privPath, err)
	}
	block, _ := pem.Decode(privPEM)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, nil, fmt.Errorf("invalid PEM in %q: expected \"PRIVATE KEY\" (PKCS#8)", privPath)
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse private key: %w", err)
	}
	privKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("key in %q is not Ed25519", privPath)
	}

	pubPEM, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read public key %q: %w", pubPath, err)
	}
	block, _ = pem.Decode(pubPEM)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, nil, fmt.Errorf("invalid PEM in %q: expected \"PUBLIC KEY\" (PKIX)", pubPath)
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse public key: %w", err)
	}
	pubKey, ok := pub.(ed25519.PublicKey)
	if !ok {
		return nil, nil, fmt.Errorf("key in %q is not Ed25519", pubPath)
	}

	return privKey, pubKey, nil
}

// initDB создаёт пул соединений и проверяет доступность БД.
func initDB(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to db: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}
	return db, nil
}
