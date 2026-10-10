package main

import (
	"log/slog"
	"os"

	"github.com/Olamilekan-12/gatekeep/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "err", err)
		os.Exit(1)
	}
	slog.Info("starting gatekeep", "addr", cfg.Addr, "shutdown_timeout", cfg.ShutdownTimeout)
}
