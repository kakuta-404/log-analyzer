package consumer

import (
	"context"
	"encoding/json"

	"github.com/kakuta-404/log-analyzer/cassandra-writer/internal/writer"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/segmentio/kafka-go"
	"golang.org/x/exp/slog"
)

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

type KafkaConsumer struct {
	reader *kafka.Reader
	writer *writer.CassandraWriter
}

func NewKafkaConsumer(cfg Config, writer *writer.CassandraWriter) (*KafkaConsumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaConsumer{
		reader: reader,
		writer: writer,
	}, nil
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	defer c.reader.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				slog.Error("error reading message", "error", err)
				continue
			}

			var event common.Event
			if err := json.Unmarshal(message.Value, &event); err != nil {
				slog.Error("error unmarshalling message", "error", err)
				continue
			}

			if err := c.writer.WriteEvent(&event); err != nil {
				slog.Error("error writing event", "error", err)
				continue
			}
		}
	}
}
