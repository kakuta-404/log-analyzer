package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gocql/gocql"
	"github.com/kakuta-404/log-analyzer/cassandra-writer/internal/consumer"
	"github.com/kakuta-404/log-analyzer/cassandra-writer/internal/writer"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cassandra"
	tcKafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/wait"
)

func createIndependentSession(host string, port string, keyspace string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(host + ":" + port)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 5 * time.Second
	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}

	return cluster.CreateSession()
}

func TestIntegrationCassandraWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start Cassandra container
	cassandraContainer, err := cassandra.RunContainer(ctx,
		testcontainers.WithImage("bitnami/cassandra:5.0.4"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Created default superuser role").
				WithStartupTimeout(60*time.Second),
		),
		testcontainers.WithEnv(map[string]string{
			"CASSANDRA_PASSWORD_SEEDER": "no",
			"CASSANDRA_AUTHENTICATOR":   "AllowAllAuthenticator",
			"CASSANDRA_AUTHORIZER":      "AllowAllAuthorizer",
		}),
	)
	require.NoError(t, err)
	defer func() {
		if err := cassandraContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	cassandraHost, err := cassandraContainer.Host(ctx)
	require.NoError(t, err)
	cassandraPort, err := cassandraContainer.MappedPort(ctx, "9042")
	require.NoError(t, err)

	// Start Kafka container
	kafkaContainer, err := tcKafka.RunContainer(ctx,
		testcontainers.WithImage("confluentinc/cp-kafka:7.3.2"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Kafka Server started").WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)
	defer func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	kafkaHost, err := kafkaContainer.Host(ctx)
	require.NoError(t, err)
	kafkaPort, err := kafkaContainer.MappedPort(ctx, "9093")
	require.NoError(t, err)

	// Initialize Cassandra writer
	var cw *writer.CassandraWriter
	for i := 0; i < 5; i++ {
		cw, err = writer.NewCassandraWriter(writer.Config{
			Hosts:    []string{cassandraHost + ":" + cassandraPort.Port()},
			Keyspace: "logs_test",
		})
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	require.NoError(t, err, "Failed to connect to Cassandra after retries")

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

	// Create Kafka producer
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{kafkaHost + ":" + kafkaPort.Port()}, config)
	require.NoError(t, err)
	defer producer.Close()

	// Wait for Kafka to be ready
	time.Sleep(10 * time.Second)

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

	// Produce test event to Kafka
	msg := &sarama.ProducerMessage{
		Topic: "logs_test",
		Value: sarama.StringEncoder(eventJSON),
	}
	_, _, err = producer.SendMessage(msg)
	require.NoError(t, err)

	// Wait for message to be processed with retries
	var count int
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)

		session, err := createIndependentSession(cassandraHost, cassandraPort.Port(), "logs_test")
		if err != nil {
			t.Logf("Error creating session: %v", err)
			continue
		}
		defer session.Close()

		err = session.Query("SELECT COUNT(*) FROM events").Scan(&count)
		if err != nil {
			t.Logf("Error querying count: %v", err)
			continue
		}

		if count == 1 {
			break
		}
	}
	assert.Equal(t, 1, count, "Expected one record in Cassandra")

	// Verify data with retries
	var storedEvent common.Event
	var found bool
	for i := 0; i < 5; i++ {
		session, err := createIndependentSession(cassandraHost, cassandraPort.Port(), "logs_test")
		if err != nil {
			t.Logf("Error creating session: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		defer session.Close()

		err = session.Query("SELECT name, project_id, timestamp, log_data FROM events LIMIT 1").
			Scan(&storedEvent.Name, &storedEvent.ProjectID, &storedEvent.EventTimestamp, &storedEvent.Log)
		if err == nil {
			found = true
			break
		}
		t.Logf("Error querying event: %v", err)
		time.Sleep(1 * time.Second)
	}

	require.True(t, found, "Failed to find stored event")

	assert.Equal(t, testEvent.Name, storedEvent.Name)
	assert.Equal(t, testEvent.ProjectID, storedEvent.ProjectID)
	assert.Equal(t, testEvent.Log, storedEvent.Log)
}
