package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/docker/go-connections/nat"
	"github.com/kakuta-404/log-analyzer/clickhouse-writer/internal/consumer"
	"github.com/kakuta-404/log-analyzer/clickhouse-writer/internal/writer"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcKafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/wait"
)

func init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func createTestConnection(host string) (clickhouse.Conn, error) {
	return clickhouse.Open(&clickhouse.Options{
		Addr: []string{host},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
	})
}

func TestIntegrationClickHouseWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start ClickHouse container
	clickhouseContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "bitnami/clickhouse:25.5.1",
			ExposedPorts: []string{"9000/tcp"},
			Env: map[string]string{
				"ALLOW_EMPTY_PASSWORD": "yes",
			},
			WaitingFor: wait.ForAll(
				wait.ForLog("Ready for connections"),
				wait.ForListeningPort("9000/tcp"),
			).WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	defer clickhouseContainer.Terminate(ctx)

	clickhouseHost, err := clickhouseContainer.Host(ctx)
	require.NoError(t, err)
	clickhousePort, err := clickhouseContainer.MappedPort(ctx, "9000")
	require.NoError(t, err)

	// Start Kafka container
	kafkaContainer, err := tcKafka.RunContainer(ctx,
		testcontainers.WithImage("confluentinc/cp-kafka:7.3.2"),
	)
	require.NoError(t, err)
	defer kafkaContainer.Terminate(ctx)

	kafkaHost, err := kafkaContainer.Host(ctx)
	require.NoError(t, err)
	kafkaPort, err := kafkaContainer.MappedPort(ctx, "9093")
	require.NoError(t, err)

	// Initialize ClickHouse writer with retries
	var cw *writer.ClickHouseWriter
	for i := 0; i < 5; i++ {
		cw, err = writer.NewClickHouseWriter(writer.Config{
			Host: clickhouseHost + ":" + clickhousePort.Port(),
		})
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	require.NoError(t, err, "Failed to connect to ClickHouse after retries")

	createKafkaTopic(t, kafkaHost, kafkaPort)

	// Initialize Kafka consumer
	kc, err := consumer.NewKafkaConsumer(consumer.Config{
		Brokers: []string{kafkaHost + ":" + kafkaPort.Port()},
		Topic:   "logs_test",
		GroupID: "test-group",
	}, cw)
	require.NoError(t, err)

	// Start consumer in background
	consumerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		err := kc.Start(consumerCtx)
		if err != nil && err != context.Canceled {
			t.Errorf("consumer error: %v", err)
		}
	}()

	// let consumer subscribe before writing the message
	time.Sleep(2 * time.Second)

	// Create Kafka writer
	w := &kafka.Writer{
		Addr:  kafka.TCP(kafkaHost + ":" + kafkaPort.Port()),
		Topic: "logs_test",
	}
	defer w.Close()

	// Wait for Kafka to be ready
	time.Sleep(5 * time.Second)

	// Create test event
	testEvent := &common.Event{
		Name:           "test_event",
		ProjectID:      "test_project",
		EventTimestamp: time.Now().UTC(),
		Log: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	// Serialize event to JSON
	eventJSON, err := json.Marshal(testEvent)
	require.NoError(t, err)

	// Write test event to Kafka
	err = w.WriteMessages(ctx, kafka.Message{
		Value: eventJSON,
	})
	require.NoError(t, err)
	slog.Info("Test message produced to kafka.")

	// Wait for message to be processed
	time.Sleep(5 * time.Second)

	// Verify data with retries
	slog.Info("Attempting to read test data from ClickHouse.")
	var count uint64
	conn, err := createTestConnection(clickhouseHost + ":" + clickhousePort.Port())
	require.NoError(t, err)
	defer conn.Close()

	for i := 0; i < 5; i++ {
		slog.Info("Querying ClickHouse to verify test.")
		err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM events").Scan(&count)
		if err == nil && count == 1 {
			break
		}
		slog.Error("Query executed.", slog.String("error", fmt.Sprintf("%v", err)), slog.Uint64("count", count))
		time.Sleep(1 * time.Second)
	}
	assert.Equal(t, uint64(1), count, "Expected one record in ClickHouse")

	// Verify event data
	var storedEvent struct {
		Name      string
		ProjectID string
		Log       map[string]string
	}

	err = conn.QueryRow(ctx, `
		SELECT name, project_id, log_data 
		FROM events 
		LIMIT 1
	`).Scan(&storedEvent.Name, &storedEvent.ProjectID, &storedEvent.Log)
	require.NoError(t, err)

	assert.Equal(t, testEvent.Name, storedEvent.Name)
	assert.Equal(t, testEvent.ProjectID, storedEvent.ProjectID)
	assert.Equal(t, testEvent.Log, storedEvent.Log)
}

// creating kafka topic and partitions beforehand will make consumer and producer work properly
// otherwise it will fallabck to the default on-demand creation that leads to malfunction
// probably due to the order of commands and configs begin executed.
func createKafkaTopic(t *testing.T, kafkaHost string, kafkaPort nat.Port) {
	slog.Info("Creating kafka topic.")
	conn, err := kafka.Dial("tcp", kafkaHost+":"+kafkaPort.Port())
	require.NoError(t, err)
	defer conn.Close()

	controller, err := conn.Controller()
	require.NoError(t, err)

	controllerConn, err := kafka.Dial("tcp", controller.Host+":"+fmt.Sprintf("%d", controller.Port))
	require.NoError(t, err)
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             "logs_test",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	require.NoError(t, err)
	slog.Info("Created topic logs_test")
}
