package cassandra

import (
	"github.com/gocql/gocql"
	"log"
)

var Session *gocql.Session

func Init(host string) error {
	cluster := gocql.NewCluster(host)
	cluster.Keyspace = "logdata"
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}

	Session = session
	log.Println("Connected to Cassandra")
	return nil
}
