package writer

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/matin/log-analyzer/common"
)

type Config struct {
	Hosts    []string
	Keyspace string
}

type CassandraWriter struct {
	session *gocql.Session
}

func NewCassandraWriter(cfg Config) (*CassandraWriter, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = time.Second * 5

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	// Create keyspace if not exists
	if err := createKeyspace(session, cfg.Keyspace); err != nil {
		return nil, err
	}

	// Create table if not exists
	if err := createTable(session); err != nil {
		return nil, err
	}

	return &CassandraWriter{
		session: session,
	}, nil
}

func (w *CassandraWriter) WriteEvent(event *common.Event) error {
	// Create composite key: ProjectID_Name_Timestamp
	key := fmt.Sprintf("%s_%s_%d",
		event.ProjectID,
		event.Name,
		event.EventTimestamp.UnixNano(),
	)

	query := `INSERT INTO events (key, project_id, name, timestamp, log_data) VALUES (?, ?, ?, ?, ?)`
	return w.session.Query(query,
		key,
		event.ProjectID,
		event.Name,
		event.EventTimestamp,
		event.Log,
	).Exec()
}

func createKeyspace(session *gocql.Session, keyspace string) error {
	query := fmt.Sprintf(`
		CREATE KEYSPACE IF NOT EXISTS %s
		WITH replication = {
			'class': 'SimpleStrategy',
			'replication_factor': 1
		}`, keyspace)

	return session.Query(query).Exec()
}

func createTable(session *gocql.Session) error {
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
