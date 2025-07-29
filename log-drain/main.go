package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kakuta-404/log-analyzer/log-drain/internal/auth"
	"github.com/kakuta-404/log-analyzer/log-drain/internal/config"
	"github.com/kakuta-404/log-analyzer/log-drain/internal/handler"
	"github.com/kakuta-404/log-analyzer/log-drain/internal/kafka"
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
	log.Println("start") 
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize services
	authService, err := auth.NewService(cfg.CockroachDB)
	if err != nil {
		slog.Error("failed to create auth service", "error", err)
		os.Exit(1)
	}

	producer, err := kafka.NewProducer(cfg.Kafka)
	if err != nil {
		slog.Error("failed to create kafka producer", "error", err)
		os.Exit(1)
	}

	log.Println("kafka has been created")

	// Setup HTTP handler
	h := handler.New(authService, producer)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go h.Start(cfg.Server.Port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	slog.Info("received shutdown signal", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	if err := h.Shutdown(shutdownCtx); err != nil {
		slog.Error("error during shutdown", "error", err)
	}

	if err := producer.Close(); err != nil {
		slog.Error("error closing kafka producer", "error", err)
	}

	slog.Info("service shutdown complete")
}
