package cassandra

import (
	"log/slog"
	"time"

	"github.com/gocql/gocql"
)

var Session *gocql.Session

func Init(host string) error {
	var err error
	maxRetries := 5
	retryDelay := time.Second * 5

	// First create a session without keyspace to create it if needed
	setupCluster := gocql.NewCluster(host)
	setupCluster.Consistency = gocql.Quorum
	setupCluster.Timeout = time.Second * 10
	setupCluster.ConnectTimeout = time.Second * 10

	setupSession, err := setupCluster.CreateSession()
	if err != nil {
		return err
	}
	defer setupSession.Close()

	// Create keyspace if not exists
	err = setupSession.Query(`
		CREATE KEYSPACE IF NOT EXISTS log_analyzer
		WITH replication = {
			'class': 'SimpleStrategy',
			'replication_factor': 1
		}`).Exec()
	if err != nil {
		return err
	}

	// Now connect with the keyspace
	for i := 0; i < maxRetries; i++ {
		cluster := gocql.NewCluster(host)
		cluster.Keyspace = "log_analyzer"
		cluster.Consistency = gocql.Quorum
		cluster.Timeout = time.Second * 10
		cluster.ConnectTimeout = time.Second * 10

		if Session, err = cluster.CreateSession(); err == nil {
			slog.Info("Connected to Cassandra")
			return nil
		}

		slog.Warn("Failed to connect to Cassandra, retrying...",
			"attempt", i+1,
			"error", err)
		time.Sleep(retryDelay)
	}

	return err
}
