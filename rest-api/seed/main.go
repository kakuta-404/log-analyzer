package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gocql/gocql"
)

func main() {
	log.Println("Waiting for databases to be ready...")
	time.Sleep(10 * time.Second)

	// ---- Cassandra Setup ----
	cluster := gocql.NewCluster("cassandra")
	cluster.Keyspace = "system" // Use temporary keyspace to create ours
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal("Cassandra connection failed:", err)
	}
	defer session.Close()

	// Create keyspace and table
	err = session.Query(`CREATE KEYSPACE IF NOT EXISTS logsystem 
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`).Exec()
	if err != nil {
		log.Fatal("Failed to create keyspace:", err)
	}

	cluster.Keyspace = "logsystem"
	session, err = cluster.CreateSession()
	if err != nil {
		log.Fatal("Failed to reconnect with keyspace:", err)
	}
	defer session.Close()

	err = session.Query(`CREATE TABLE IF NOT EXISTS events (
		project_id text,
		name text,
		timestamp timestamp,
		log_data map<text, text>,
		PRIMARY KEY ((project_id, name), timestamp)
	)`).Exec()
	if err != nil {
		log.Fatal("Failed to create Cassandra table:", err)
	}

	if err := seedCassandra(session, "/seed/data.csv"); err != nil {
		log.Fatal("Failed seeding Cassandra:", err)
	}

	// ---- ClickHouse Setup ----
	clickConn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"clickhouse:9000"},
		Auth: clickhouse.Auth{Database: "default"},
	})
	if err != nil {
		log.Fatal("ClickHouse connection failed:", err)
	}

	if err := seedClickHouse(clickConn, "/seed/data.csv"); err != nil {
		log.Fatal("Failed seeding ClickHouse:", err)
	}

	log.Println("✅ Seeding complete.")
}

func seedCassandra(session *gocql.Session, path string) error {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return err
	}

	for _, row := range rows[1:] {
		timestamp, _ := time.Parse("2006-01-02 15:04:05", row[2])
		var logMap map[string]string
		_ = json.Unmarshal([]byte(row[3]), &logMap)

		if err := session.Query(`
			INSERT INTO events (project_id, name, timestamp, log_data)
			VALUES (?, ?, ?, ?)`,
			row[0], row[1], timestamp, logMap,
		).Exec(); err != nil {
			log.Println("Cassandra insert error:", err)
		}
	}
	return nil
}

func seedClickHouse(conn clickhouse.Conn, path string) error {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return err
	}

	ctx := context.Background()
	batch, err := conn.PrepareBatch(ctx, `
		INSERT INTO events (project_id, name, timestamp, log_data)
	`)
	if err != nil {
		return err
	}

	for _, row := range rows[1:] {
		timestamp, _ := time.Parse("2006-01-02 15:04:05", row[2])
		var logMap map[string]string
		_ = json.Unmarshal([]byte(row[3]), &logMap)

		if err := batch.Append(row[0], row[1], timestamp, logMap); err != nil {
			log.Println("ClickHouse insert error:", err)
		}
	}
	return batch.Send()
}
