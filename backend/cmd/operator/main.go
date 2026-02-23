package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/portwhine/portwhine/internal/operator"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	cfg, err := operator.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Setup structured logging
	var level slog.Level
	switch cfg.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	slog.Info("starting portwhine operator",
		"grpc_addr", cfg.Server.GRPCAddr,
		"runtime", cfg.Runtime.Type,
	)

	op, err := operator.New(cfg, logger)
	if err != nil {
		slog.Error("failed to create operator", "error", err)
		os.Exit(1)
	}

	if err := op.Run(); err != nil {
		slog.Error("operator exited with error", "error", err)
		os.Exit(1)
	}
}
