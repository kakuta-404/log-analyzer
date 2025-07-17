package kafka

import (
	"context"
	"encoding/json"

	"github.com/kakuta-404/log-analyzer/common"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(cfg struct {
	Brokers []string
	Topic   string
}) (*Producer, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
	})

	return &Producer{
		writer: writer,
	}, nil
}

func (p *Producer) SendEvent(ctx context.Context, event *common.Event) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Value: value,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
