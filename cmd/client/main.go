package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/avc-dev/gophkeeper/internal/client/command"
)

// Version и BuildTime устанавливаются через ldflags при сборке.
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	root := command.NewRootCmd(Version, BuildTime)
	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
