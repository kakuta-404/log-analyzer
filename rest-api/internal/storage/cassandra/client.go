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

	for i := 0; i < maxRetries; i++ {
		cluster := gocql.NewCluster(host)
		cluster.Keyspace = "log_analyzer" // Match keyspace from seed service
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
