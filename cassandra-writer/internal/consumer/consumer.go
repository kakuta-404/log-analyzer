package consumer

import (
	"context"
	"encoding/json"
	"log"

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
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumer: group,
		writer:   writer,
		topic:    cfg.Topic,
	}, nil
}

func (k *KafkaConsumer) Start(ctx context.Context) error {
	for {
		if err := k.consumer.Consume(ctx, []string{k.topic}, k); err != nil {
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (k *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event common.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		if err := k.writer.WriteEvent(&event); err != nil {
			log.Printf("Error writing event: %v", err)
			continue
		}

		session.MarkMessage(msg, "")
	}
	return nil
}

func (k *KafkaConsumer) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (k *KafkaConsumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
