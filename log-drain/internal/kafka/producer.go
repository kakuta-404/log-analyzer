package kafka

import (
	"context"
	"encoding/json"
	"log"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/segmentio/kafka-go"
	"fmt"
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

	CreateTopic()

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

// creating topics for 500 for fixing internal error

func CreateTopic() {
	conn, err := kafka.Dial("tcp", "kafka:9092")
    if err != nil {
        log.Fatalf("failed to connect to Kafka broker: %v", err)
    }
    defer conn.Close()

    controller, err := conn.Controller()
    if err != nil {
        log.Fatalf("failed to get controller: %v", err)
    }

    controllerConn, err := kafka.Dial("tcp", controller.Host+":"+fmt.Sprint(controller.Port))
    if err != nil {
        log.Fatalf("failed to connect to controller: %v", err)
    }
    defer controllerConn.Close()

    topic := "logs" 
    partitions := 1
    replication := 1

    err = controllerConn.CreateTopics(kafka.TopicConfig{
        Topic:             topic,
        NumPartitions:     partitions,
        ReplicationFactor: replication,
    })

    if err != nil {
        log.Fatalf("failed to create topic: %v", err)
    }

    log.Printf("topic %s created successfully", topic)
}
