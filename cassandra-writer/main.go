package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/matin/log-analyzer/cassandra-writer/internal/consumer"
	"github.com/matin/log-analyzer/cassandra-writer/internal/writer"
)

func main() {
	log.Println("Starting Cassandra Writer Service...")

	// Initialize Cassandra writer
	cw, err := writer.NewCassandraWriter(writer.Config{
		Hosts:    []string{"cassandra:9042"},
		Keyspace: "logs",
	})
	if err != nil {
		log.Fatalf("Failed to create Cassandra writer: %v", err)
	}

	// Initialize Kafka consumer
	kc, err := consumer.NewKafkaConsumer(consumer.Config{
		Brokers: []string{"kafka:9092"},
		Topic:   "logs",
		GroupID: "cassandra-writer",
	}, cw)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	if err := kc.Start(ctx); err != nil {
		log.Fatalf("Error running consumer: %v", err)
	}
}
