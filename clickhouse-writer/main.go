package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	 _ "github.com/lib/pq"
	"github.com/kakuta-404/log-analyzer/clickhouse-writer/internal/consumer"
	"github.com/kakuta-404/log-analyzer/clickhouse-writer/internal/writer"
	"github.com/kakuta-404/log-analyzer/common"
)

func init() {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func main() {
	
	slog.Info("starting clickhouse writer service")
	
	slog.Info("initializing clickhouse writer...")
	// Initialize ClickHouse writer
	cw, err := writer.NewClickHouseWriter(writer.Config{
		Host: "clickhouse:9000",
	})
	
	if err != nil {
		slog.Error("failed to create clickhouse writer", "error", err)
		os.Exit(1)
	}
	slog.Info("successfully connected to clickhouse")

	slog.Info("initializing kafka consumer...")
	// Initialize Kafka consumer
	kc, err := consumer.NewKafkaConsumer(consumer.Config{
		Brokers: common.KafkaBrokers,
		Topic:   "logs",
		GroupID: "clickhouse-writer",
	}, cw)
	if err != nil {
		slog.Error("failed to create kafka consumer", "error", err)
		os.Exit(1)
	}
	slog.Info("successfully connected to kafka")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		slog.Info("received shutdown signal", "signal", sig)
		cancel()
	}()

	slog.Info("starting to consume messages...")
	if err := kc.Start(ctx); err != nil {
		slog.Error("error running consumer", "error", err)
		os.Exit(1)
	}
	slog.Info("clickhouse writer service shutdown complete")
}
