package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/handlers"
	cassandraStore "github.com/kakuta-404/log-analyzer/rest-api/internal/storage/cassandra"
	clickhouseStore "github.com/kakuta-404/log-analyzer/rest-api/internal/storage/clickhouse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIntegrationRESTAPI(t *testing.T) {
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

	// Get connection details
	cassandraHost, err := cassandraContainer.Host(ctx)
	require.NoError(t, err)
	cassandraPort, err := cassandraContainer.MappedPort(ctx, "9042")
	require.NoError(t, err)

	clickhouseHost, err := clickhouseContainer.Host(ctx)
	require.NoError(t, err)
	clickhousePort, err := clickhouseContainer.MappedPort(ctx, "9000")
	require.NoError(t, err)

	// Initialize databases
	err = SetupCassandra(fmt.Sprintf("%s:%s", cassandraHost, cassandraPort.Port()))
	require.NoError(t, err)

	err = SetupClickHouse(fmt.Sprintf("%s:%s", clickhouseHost, clickhousePort.Port()))
	require.NoError(t, err)

	// Initialize Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/projects/:id/events", handlers.GetGroupedEvents)
	router.GET("/projects/:id/search", handlers.SearchGroupedEvents)

	// Test data
	testEvent := common.Event{
		ProjectID:      "test_project",
		Name:           "test_event",
		EventTimestamp: time.Now(),
		Log: map[string]string{
			"user_id": "123",
			"action":  "click",
		},
	}

	// Insert test data
	err = InsertTestData(cassandraHost+":"+cassandraPort.Port(), clickhouseHost+":"+clickhousePort.Port(), testEvent)
	require.NoError(t, err)

	// Test GET /projects/:id/events
	t.Run("GetGroupedEvents", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/projects/test_project/events", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		groups, ok := response["groups"].([]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, groups)
	})

	// Test GET /projects/:id/search
	t.Run("SearchGroupedEvents", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/projects/test_project/search?user_id=123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		groups, ok := response["groups"].([]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, groups)
	})
}

func SetupCassandra(host string) error {
	cluster := gocql.NewCluster(host)
	cluster.Keyspace = "system"
	cluster.Timeout = time.Second * 10

	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}
	defer session.Close()

	err = session.Query(`
		CREATE KEYSPACE IF NOT EXISTS logdata
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
	`).Exec()
	if err != nil {
		return err
	}

	cluster.Keyspace = "logdata"
	session, err = cluster.CreateSession()
	if err != nil {
		return err
	}
	defer session.Close()

	return session.Query(`
		CREATE TABLE IF NOT EXISTS events_by_name (
			project_id text,
			name text,
			timestamp timestamp,
			log_data map<text, text>,
			PRIMARY KEY ((project_id, name), timestamp)
		)
	`).Exec()
}

func SetupClickHouse(host string) error {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{host},
		Auth: clickhouse.Auth{
			Database: "default",
		},
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS events (
			project_id String,
			name String,
			timestamp DateTime,
			log_data Map(String, String)
		) ENGINE = MergeTree()
		ORDER BY (project_id, name, timestamp)
	`)
}

func InsertTestData(cassandraHost, clickhouseHost string, event common.Event) error {
	// Insert into Cassandra
	err := cassandraStore.Init(cassandraHost)
	if err != nil {
		return err
	}

	err = cassandraStore.Session.Query(`
		INSERT INTO events_by_name (project_id, name, timestamp, log_data)
		VALUES (?, ?, ?, ?)
	`, event.ProjectID, event.Name, event.EventTimestamp, event.Log).Exec()
	if err != nil {
		return err
	}

	// Insert into ClickHouse
	err = clickhouseStore.Init(clickhouseHost)
	if err != nil {
		return err
	}

	return clickhouseStore.Conn.Exec(context.Background(), `
		INSERT INTO events (project_id, name, timestamp, log_data)
		VALUES (?, ?, ?, ?)
	`, event.ProjectID, event.Name, event.EventTimestamp, event.Log)
}
