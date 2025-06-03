package main

import (
	"context"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gocql/gocql"
	"github.com/matin/log-analyzer/cassandra-writer/internal/consumer"
	"github.com/matin/log-analyzer/cassandra-writer/internal/writer"
	"github.com/matin/log-analyzer/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cassandra"
	tcKafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIntegrationCassandraWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start Cassandra container
	cassandraContainer, err := cassandra.RunContainer(ctx,
		testcontainers.WithImage("cassandra:latest"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Created default superuser role").
				WithStartupTimeout(60*time.Second),
		),
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
		testcontainers.WithImage("confluentinc/cp-kafka:latest"),
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
	cw, err := writer.NewCassandraWriter(writer.Config{
		Hosts:    []string{cassandraHost + ":" + cassandraPort.Port()},
		Keyspace: "logs_test",
	})
	require.NoError(t, err)

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

	// Create test event
	testEvent := &common.Event{
		Name:           "test_event",
		ProjectID:      "test_project",
		EventTimestamp: time.Now(),
		Log: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	// Produce test event to Kafka
	msg := &sarama.ProducerMessage{
		Topic: "logs_test",
		Value: sarama.StringEncoder(testEvent.Name),
	}
	_, _, err = producer.SendMessage(msg)
	require.NoError(t, err)

	// Wait for message to be processed
	time.Sleep(5 * time.Second)

	// Verify data in Cassandra
	cluster := gocql.NewCluster(cassandraHost + ":" + cassandraPort.Port())
	cluster.Keyspace = "logs_test"
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	require.NoError(t, err)
	defer session.Close()

	var count int
	err = session.Query("SELECT COUNT(*) FROM events").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Expected one record in Cassandra")

	var storedEvent common.Event
	err = session.Query("SELECT name, project_id, timestamp, log_data FROM events LIMIT 1").
		Scan(&storedEvent.Name, &storedEvent.ProjectID, &storedEvent.EventTimestamp, &storedEvent.Log)
	require.NoError(t, err)

	assert.Equal(t, testEvent.Name, storedEvent.Name)
	assert.Equal(t, testEvent.ProjectID, storedEvent.ProjectID)
	assert.Equal(t, testEvent.Log, storedEvent.Log)
}
