package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kakuta-404/log-analyzer/cassandra-writer/internal/consumer"
	"github.com/kakuta-404/log-analyzer/cassandra-writer/internal/writer"
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
	slog.Info("starting cassandra writer service")

	slog.Info("initializing cassandra writer...")
	// Initialize Cassandra writer
	cw, err := writer.NewCassandraWriter(writer.Config{
		Hosts:    []string{"cassandra:9042"},
		Keyspace: "logs",
	})
	if err != nil {
		slog.Error("failed to create cassandra writer", "error", err)
		os.Exit(1)
	}
	slog.Info("successfully connected to cassandra")

	slog.Info("initializing kafka consumer...")
	// Initialize Kafka consumer
	kc, err := consumer.NewKafkaConsumer(consumer.Config{
		Brokers: []string{"kafka:9092"},
		Topic:   "logs",
		GroupID: "cassandra-writer",
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
	slog.Info("cassandra writer service shutdown complete")
}
