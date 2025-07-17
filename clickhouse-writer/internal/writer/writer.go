package writer

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"

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


var ProjectIDs []int 

func AddProjectID(projectID int) {
    needForMake := false
	for  i := 0 ; i < len(ProjectIDs) ; i++ {
		if (projectID != ProjectIDs[i]) {
			needForMake = true
			ProjectIDs = append(ProjectIDs, projectID)
			break
		}
	}

	if needForMake {
		
	}

}

var ClickhouseConnection driver.Conn

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

var Db *sql.DB


func ConnectToCockroachdb() {
	var err error
	connStr := "postgresql://root@localhost:26257/defaultdb?sslmode=disable"
	Db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("can not connect to databse -- > " , err)
	} else {
		log.Printf("connected to the database")
	}
}

func GetSreachableKeys(projectID int) ([]string, error) {
	var searchable string

    err := Db.QueryRow(
        "SELECT searchable_keys FROM projects WHERE id = $1", projectID,
    ).Scan(&searchable)

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("not found")
    }
    if err != nil {
        return nil, err
    }

    var keys [] string
    err = json.Unmarshal([]byte(searchable), &keys)
    if err != nil {
        return nil, fmt.Errorf("invalid")
    }

    return keys, nil
}

func createTable(conn clickhouse.Conn, projectID int) error {
	searchableKeys, err := GetSreachableKeys(projectID)
	if  err != nil {
		return err
	}
	slog.Info("creating table for the projectID")
	tableName := fmt.Sprintf("events_%d", projectID)
    CreationQuery := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            name String,
            timestamp DateTime,
            log_data Map(String, String)
        ) ENGINE = MergeTree()
        ORDER BY timestamp
    `, tableName)
	
	if err := conn.Exec(context.Background(), CreationQuery); err != nil {
        return fmt.Errorf("failed to create table %s: %w", tableName, err)
    }

	indexSize := len(searchableKeys)
    if indexSize == 0 {
        indexSize = 100 // Default fallback if no keys
    }
    alterQuery := fmt.Sprintf(`
        ALTER TABLE %s
        ADD INDEX IF NOT EXISTS searchable_keys_index (mapKeys(log_data)) TYPE set(%d) GRANULARITY 1
    `, tableName, indexSize)
    
    if err := conn.Exec(context.Background(), alterQuery); err != nil {
        return fmt.Errorf("failed to add index to table %s: %w", tableName, err)
    }
    
    bloomQuery := fmt.Sprintf(`
        ALTER TABLE %s
        ADD INDEX IF NOT EXISTS searchable_keys_bloom (mapKeys(log_data)) TYPE bloom_filter(0.01) GRANULARITY 1
    `, tableName)
    
    if err := conn.Exec(context.Background(), bloomQuery); err != nil {
        return fmt.Errorf("failed to add bloom filter index to table %s: %w", tableName, err)
    }

    optimizeQuery := fmt.Sprintf(`OPTIMIZE TABLE %s FINAL`, tableName)
    if err := conn.Exec(context.Background(), optimizeQuery); err != nil {
        slog.Warn("failed to optimize table, indexes may not apply immediately", "tableName", tableName, "error", err)
    }

	return nil
}
