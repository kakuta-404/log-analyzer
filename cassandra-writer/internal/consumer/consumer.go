package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Shopify/sarama"
	"github.com/kakuta-404/log-analyzer/cassandra-writer/internal/writer"
	"github.com/kakuta-404/log-analyzer/common"
)

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	writer   *writer.CassandraWriter
	topic    string
}

func NewKafkaConsumer(cfg Config, writer *writer.CassandraWriter) (*KafkaConsumer, error) {
	slog.Debug("configuring kafka consumer",
		"brokers", cfg.Brokers,
		"topic", cfg.Topic,
		"group", cfg.GroupID)

	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %v", err)
	}

	slog.Info("successfully created kafka consumer group")

	return &KafkaConsumer{
		consumer: consumer,
		writer:   writer,
		topic:    cfg.Topic,
	}, nil
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	slog.Info("starting to consume from topic", "topic", c.topic)

	for {
		select {
		case <-ctx.Done():
			slog.Info("context cancelled, stopping consumer")
			return c.consumer.Close()
		default:
			err := c.consumer.Consume(ctx, []string{c.topic}, c)
			if err != nil {
				return fmt.Errorf("error from consumer: %v", err)
			}
		}
	}
}

func (c *KafkaConsumer) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (c *KafkaConsumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (c *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			var event common.Event
			if err := json.Unmarshal(message.Value, &event); err != nil {
				slog.Error("error unmarshaling message",
					"error", err,
					"value", string(message.Value))
				session.MarkMessage(message, "")
				continue
			}

			if err := c.writer.WriteEvent(&event); err != nil {
				slog.Error("error writing event",
					"error", err,
					"event", event)
				continue
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
