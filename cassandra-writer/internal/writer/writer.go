package writer

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/kakuta-404/log-analyzer/common"
	"golang.org/x/exp/slog"
)

type Config struct {
	Host string
}

type CassandraWriter struct {
	session *gocql.Session
}

func NewCassandraWriter(cfg Config) (*CassandraWriter, error) {
	slog.Info("Initializing Cassandra writer.")

	cluster := gocql.NewCluster(cfg.Host)
	cluster.Keyspace = "log_analyzer"
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = time.Second * 15

	// Attempt to create keyspace
	setupSession, err := gocql.NewCluster(cfg.Host).CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create setup session: %v", err)
	}
	defer setupSession.Close()

	err = createKeyspace(setupSession)
	if err != nil {
		return nil, fmt.Errorf("failed to create keyspace: %v", err)
	}

	// Create main session with keyspace
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cassandra: %v", err)
	}

	// Create table
	if err := createTable(session); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	slog.Info("Connection to Cassandra successful.")
	return &CassandraWriter{session: session}, nil
}

func (w *CassandraWriter) WriteEvent(event *common.Event) error {
	slog.Info("writing event",
		"project_id", event.ProjectID,
		"name", event.Name,
		"timestamp", event.EventTimestamp)

	query := `INSERT INTO events (project_id, name, event_timestamp, log_data) VALUES (?, ?, ?, ?)`

	if err := w.session.Query(query,
		event.ProjectID,
		event.Name,
		event.EventTimestamp,
		event.Log,
	).Exec(); err != nil {
		slog.Error("error writing event to cassandra",
			"error", err,
			"project_id", event.ProjectID,
			"name", event.Name)
		return err
	}

	slog.Info("successfully wrote event", "project_id", event.ProjectID, "name", event.Name)
	return nil
}

func createKeyspace(session *gocql.Session) error {
	return session.Query(`
		CREATE KEYSPACE IF NOT EXISTS log_analyzer
		WITH replication = {
			'class': 'SimpleStrategy',
			'replication_factor': 1
		}`).Exec()
}

func createTable(session *gocql.Session) error {
	slog.Info("creating events table if not exists")
	return session.Query(`
		CREATE TABLE IF NOT EXISTS events (
			project_id text,
			name text,
			event_timestamp timestamp,
			log_data map<text, text>,
			PRIMARY KEY ((project_id, name), event_timestamp)
		) WITH CLUSTERING ORDER BY (event_timestamp DESC)
	`).Exec()
}
