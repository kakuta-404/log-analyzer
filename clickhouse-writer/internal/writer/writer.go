package writer

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/kakuta-404/log-analyzer/common"
	"golang.org/x/exp/slog"
)

type Config struct {
	Host string
}

type ClickHouseWriter struct {
	conn clickhouse.Conn
}

func NewClickHouseWriter(cfg Config) (*ClickHouseWriter, error) {
	slog.Info("Initializing ClickHouse writer.")
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.Host},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
		Debug: true,
	})
	if err != nil {
		slog.Error("failed to connect to clickhouse", "error", err)
		return nil, fmt.Errorf("failed to connect to clickhouse: %v", err)
	}
	slog.Info("Connection to ClickHouse successful.")

	// Create table if not exists
	if err := createTable(conn); err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}
	slog.Info("Created table successfully.")
	return &ClickHouseWriter{conn: conn}, nil
}

func (w *ClickHouseWriter) WriteEvent(event *common.Event) error {
	slog.Info("writing event",
		"project_id", event.ProjectID,
		"name", event.Name,
		"timestamp", event.EventTimestamp)

	query := `
		INSERT INTO events (project_id, name, timestamp, log_data)
		VALUES (?, ?, ?, ?)
	`
	err := w.conn.Exec(context.Background(), query,
		event.ProjectID,
		event.Name,
		event.EventTimestamp,
		event.Log,
	)
	if err != nil {
		slog.Error("error writing event to clickhouse",
			"error", err,
			"project_id", event.ProjectID,
			"name", event.Name)
		return err
	}

	slog.Info("successfully wrote event", "project_id", event.ProjectID, "name", event.Name)
	return nil
}

var db *sql.DB

func connectToCockroachdb() {
	connStr := "postgresql://root@localhost:26257/defaultdb?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("could not connect to cockraochDb", err)
		return err
	}
}

func getSreachableKeys() ([]string, error) {

}

func createTable(conn clickhouse.Conn) error {
	slog.Info("creating events table if not exists")
	query := `
		CREATE TABLE IF NOT EXISTS events (
			project_id String,
			name String,
			timestamp DateTime,
			log_data Map(String, String)
		) ENGINE = MergeTree()
		ORDER BY timestamp
	`
	return conn.Exec(context.Background(), query)
}
