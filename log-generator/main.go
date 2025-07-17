package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kakuta-404/log-analyzer/log-generator/internal/config"
	"github.com/kakuta-404/log-analyzer/log-generator/internal/generator"
	"github.com/kakuta-404/log-analyzer/log-generator/internal/handler"
	"github.com/kakuta-404/log-analyzer/log-generator/internal/projects"
)

func init() {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	projectSvc, err := projects.NewService(cfg.CockroachDB)
	if err != nil {
		slog.Error("failed to create project service", "error", err)
		os.Exit(1)
	}

	gen := generator.New(struct{ LogDrainURL string }{LogDrainURL: cfg.Generator.LogDrainURL}, projectSvc)
	h := handler.New(gen)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start periodic log generation
	ticker := time.NewTicker(cfg.Generator.Interval)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := gen.GenerateAndSend(); err != nil {
					slog.Error("failed to generate/send log", "error", err)
				}
			}
		}
	}()

	// Start HTTP server
	go h.Start(cfg.Server.Port)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	slog.Info("received shutdown signal", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()
	if err := h.Shutdown(shutdownCtx); err != nil {
		slog.Error("error during shutdown", "error", err)
	}
	slog.Info("service shutdown complete")
}
