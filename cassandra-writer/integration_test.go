package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/gocql/gocql"
	"github.com/kakuta-404/log-analyzer/cassandra-writer/internal/consumer"
	"github.com/kakuta-404/log-analyzer/cassandra-writer/internal/writer"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcKafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

func init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func createTestSession(host string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(host)
	cluster.Keyspace = "log_analyzer"
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = time.Second * 5
	return cluster.CreateSession()
}

func TestIntegrationCassandraWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start Cassandra container
	cassandraContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "bitnami/cassandra:5.0.4",
			ExposedPorts: []string{"9042/tcp"},
			Env: map[string]string{
				"CASSANDRA_PASSWORD_SEEDER": "no",
				"CASSANDRA_AUTHENTICATOR":   "AllowAllAuthenticator",
				"CASSANDRA_AUTHORIZER":      "AllowAllAuthorizer",
			},
		},
		Started: true,
	})
	require.NoError(t, err)
	defer cassandraContainer.Terminate(ctx)
	time.Sleep(150 * time.Second)
	cassandraHost, err := cassandraContainer.Host(ctx)
	require.NoError(t, err)
	cassandraPort, err := cassandraContainer.MappedPort(ctx, "9042")
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

	// Initialize Cassandra writer with retries
	var cw *writer.CassandraWriter
	for i := 0; i < 5; i++ {
		cw, err = writer.NewCassandraWriter(writer.Config{
			Host: cassandraHost + ":" + cassandraPort.Port(),
		})
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	require.NoError(t, err, "Failed to connect to Cassandra after retries")

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
	time.Sleep(15 * time.Second)

	// Verify data with retries
	slog.Info("Attempting to read test data from Cassandra.")
	session, err := createTestSession(cassandraHost + ":" + cassandraPort.Port())
	require.NoError(t, err)
	defer session.Close()

	var count int
	for i := 0; i < 5; i++ {
		err = session.Query(`SELECT COUNT(*) FROM events WHERE project_id = ? AND name = ?`,
			testEvent.ProjectID, testEvent.Name).Scan(&count)
		if err == nil && count == 1 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	assert.Equal(t, 1, count, "Expected one record in Cassandra")

	// Verify event data
	var (
		name      string
		projectID string
		logData   map[string]string
	)

	err = session.Query(`
		SELECT name, project_id, log_data 
		FROM events 
		WHERE project_id = ? AND name = ? 
		LIMIT 1`,
		testEvent.ProjectID, testEvent.Name,
	).Scan(&name, &projectID, &logData)
	require.NoError(t, err)

	assert.Equal(t, testEvent.Name, name)
	assert.Equal(t, testEvent.ProjectID, projectID)
	assert.Equal(t, testEvent.Log, logData)
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
