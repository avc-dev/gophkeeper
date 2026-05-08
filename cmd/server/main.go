package main

import (
	"log/slog"
	"os"

	"github.com/avc-dev/gophkeeper/internal/server/app"
	"github.com/avc-dev/gophkeeper/internal/server/config"
)

// Version и BuildTime подставляются через ldflags при сборке.
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	log.Info("starting server", "version", Version, "built", BuildTime)

	cfg, err := config.Load()
	if err != nil {
		log.Error("invalid config", "err", err)
		os.Exit(1)
	}

	if err := app.Run(cfg, log); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
