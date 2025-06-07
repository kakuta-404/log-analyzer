package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/kakuta-404/log-analyzer/clickhouse-writer/internal/writer"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

type KafkaConsumer struct {
	reader *kafka.Reader
	writer *writer.ClickHouseWriter
}

func NewKafkaConsumer(cfg Config, writer *writer.ClickHouseWriter) (*KafkaConsumer, error) {
	slog.Info("configuring kafka consumer",
		"brokers", cfg.Brokers,
		"topic", cfg.Topic,
		"group", cfg.GroupID)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
		GroupID: cfg.GroupID,
	})

	// Test Kafka connectivity with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	slog.Info("checking kafka connectivity")

	_, err := reader.FetchMessage(ctx)
	if err != nil && err != context.DeadlineExceeded {
		slog.Error("failed to connect to kafka",
			"error", err,
			"brokers", cfg.Brokers,
			"topic", cfg.Topic,
			"group", cfg.GroupID)
		reader.Close()
		return nil, fmt.Errorf("kafka connection check failed: %w", err)
	}

	slog.Info("kafka connectivity check successful")
	slog.Info("successfully created kafka consumer")

	return &KafkaConsumer{
		reader: reader,
		writer: writer,
	}, nil
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	slog.Info("starting to consume from topic", "topic", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			slog.Info("context cancelled, stopping consumer")
			if err := c.reader.Close(); err != nil {
				slog.Error("error closing reader", "error", err)
				return err
			}
			slog.Info("reader closed successfully")
			return nil

		default:
			slog.Info("attempting to read message")
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				// this is added for testing purposes. we needed more time to check clickhouse.
				time.Sleep(60 * time.Second)
				slog.Error("error reading message", "error", err)
				return fmt.Errorf("error reading message: %v", err)
			}
			slog.Info("message read successfully",
				"offset", msg.Offset,
				"partition", msg.Partition,
				"key", string(msg.Key))

			var event common.Event
			slog.Info("attempting to unmarshal message")
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				slog.Error("error unmarshaling message",
					"error", err,
					"value", string(msg.Value))
				continue
			}
			slog.Info("message unmarshaled successfully", "event", event)

			slog.Info("attempting to write event")
			if err := c.writer.WriteEvent(&event); err != nil {
				slog.Error("error writing event",
					"error", err,
					"event", event)
				continue
			}
			slog.Info("event written successfully", "event", event)
		}
	}
}
