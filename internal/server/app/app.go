package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os/signal"
	"syscall"

	"github.com/avc-dev/gophkeeper/internal/server/config"
	"github.com/avc-dev/gophkeeper/internal/server/handler"
	authhandler "github.com/avc-dev/gophkeeper/internal/server/handler/auth"
	authsvc "github.com/avc-dev/gophkeeper/internal/server/service/auth"
	userstore "github.com/avc-dev/gophkeeper/internal/server/storage/user"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
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

	users := userstore.New(db)
	auth := authsvc.New(users, cfg.JWTSecret)

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(handler.AuthInterceptor(auth)),
	)
	pb.RegisterAuthServiceServer(srv, authhandler.New(auth))

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
