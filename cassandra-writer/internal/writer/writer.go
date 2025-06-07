package writer

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/kakuta-404/log-analyzer/common"
	"golang.org/x/exp/slog"
)

type Config struct {
	Hosts    []string
	Keyspace string
}

type CassandraWriter struct {
	session *gocql.Session
}

func NewCassandraWriter(cfg Config) (*CassandraWriter, error) {
	// First create a cluster to create keyspace
	initCluster := gocql.NewCluster(cfg.Hosts...)
	initCluster.Consistency = gocql.Quorum
	initCluster.Timeout = time.Second * 5

	initSession, err := initCluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create init session: %v", err)
	}
	defer initSession.Close()

	// Create keyspace if not exists
	if err := createKeyspace(initSession, cfg.Keyspace); err != nil {
		return nil, fmt.Errorf("failed to create keyspace: %v", err)
	}

	// Now create the real session with the keyspace
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = time.Second * 5
	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	// Create table if not exists
	if err := createTable(session); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	return &CassandraWriter{
		session: session,
	}, nil
}

func (w *CassandraWriter) WriteEvent(event *common.Event) error {
	slog.Debug("writing event",
		"project_id", event.ProjectID,
		"name", event.Name,
		"timestamp", event.EventTimestamp)

	// Create composite key: ProjectID_Name_Timestamp
	key := fmt.Sprintf("%s_%s_%d",
		event.ProjectID,
		event.Name,
		event.EventTimestamp.UnixNano(),
	)

	query := `INSERT INTO events (key, project_id, name, timestamp, log_data) VALUES (?, ?, ?, ?, ?)`
	err := w.session.Query(query,
		key,
		event.ProjectID,
		event.Name,
		event.EventTimestamp,
		event.Log,
	).Exec()

	if err != nil {
		slog.Error("error writing event to cassandra",
			"error", err,
			"key", key)
		return err
	}

	slog.Debug("successfully wrote event", "key", key)
	return nil
}

func createKeyspace(session *gocql.Session, keyspace string) error {
	slog.Debug("creating keyspace if not exists", "keyspace", keyspace)
	query := fmt.Sprintf(`
		CREATE KEYSPACE IF NOT EXISTS %s
		WITH replication = {
			'class': 'SimpleStrategy',
			'replication_factor': 1
		}`, keyspace)

	return session.Query(query).Exec()
}

func createTable(session *gocql.Session) error {
	slog.Debug("creating events table if not exists")
	query := `
		CREATE TABLE IF NOT EXISTS events (
			key text PRIMARY KEY,
			project_id text,
			name text,
			timestamp timestamp,
			log_data map<text, text>
		)`

	return session.Query(query).Exec()
}
