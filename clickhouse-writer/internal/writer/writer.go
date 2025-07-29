package writer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	
	"regexp"
	"github.com/ClickHouse/clickhouse-go/v2"
	chdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/kakuta-404/log-analyzer/common"
	"golang.org/x/exp/slog"
)

type Config struct {
	Host string
}

type ClickHouseWriter struct {
	conn clickhouse.Conn
}


var ProjectIDs []string

func AddProjectID(projectID string) error {
	for _, id := range ProjectIDs { 
        if id == projectID {
            return nil 
        }
    }
   
    ProjectIDs = append(ProjectIDs, projectID)
    return createTable(ClickhouseConnection, projectID)
}

var ClickhouseConnection chdriver.Conn

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
	
	ClickhouseConnection = conn

	if err != nil {
		slog.Error("failed to connect to clickhouse", "error", err)
		return nil, fmt.Errorf("failed to connect to clickhouse: %v", err)
	}
	slog.Info("Connection to ClickHouse successful.")

	slog.Info("Created table successfully.")

	ConnectToCockroachdb()

	if err := GetProjectsID(); err != nil {
		slog.Error("getting projects ID faced error -- > " , err)
	}

	StartUP(conn)

	printProject()
	
	return &ClickHouseWriter{conn: conn}, nil
}

func (w *ClickHouseWriter) WriteEvent(event *common.Event) error {
	
	AddProjectID(event.ProjectID)
	slog.Info("writing event",
		"project_id", event.ProjectID,
		"name", event.Name,
		"timestamp", event.EventTimestamp)

	insertQuery := fmt.Sprintf(`INSERT INTO %s (name, timestamp, log_data) VALUES (?, ?, ?)`, event.ProjectID)

	err := w.conn.Exec(context.Background(), insertQuery,
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

var database *sql.DB


func ConnectToCockroachdb() {
	var err error
	connStr := "postgresql://root@cockroachdb:26257/defaultdb?sslmode=disable"
	database, err = sql.Open("postgres", connStr)
	if err != nil {
		slog.Error("can not connect to databse -- > " , err)
	} else {
		log.Printf("connected to the database to cockroach database")
	}
}

func GetSearchableKeys(projectID string) ([]string, error) {
	var searchable string

    err := database.QueryRow(
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

func createTable(conn clickhouse.Conn, projectID string) error {
	if database == nil {
		ConnectToCockroachdb()
	}
	searchableKeys, err := GetSearchableKeys(projectID)
	if err != nil {
		return err
	}
	slog.Info("creating table for the projectID")
	CreationQuery := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            name String,
            timestamp DateTime,
            log_data Map(String, String)
        ) ENGINE = MergeTree()
        ORDER BY timestamp
    `, projectID)

	if err := conn.Exec(context.Background(), CreationQuery); err != nil {
		return fmt.Errorf("failed to create table %s: %w", projectID, err)
	}


	keyRegexp := regexp.MustCompile(`[^\w\d_]`) 
	for _, key := range searchableKeys {
		safeKey := keyRegexp.ReplaceAllString(key, "_")
		indexName := fmt.Sprintf("idx_%s", safeKey)
		alterIdxQuery := fmt.Sprintf(`
			ALTER TABLE %s
			ADD INDEX IF NOT EXISTS %s (log_data['%s']) TYPE bloom_filter(0.01) GRANULARITY 1
		`, projectID, indexName, key)
		if err := conn.Exec(context.Background(), alterIdxQuery); err != nil {
			return fmt.Errorf("failed to add bloom filter index for key %s to table %s: %w", key, projectID, err)
		}
	}

	optimizeQuery := fmt.Sprintf(`OPTIMIZE TABLE %s FINAL`, projectID)
	if err := conn.Exec(context.Background(), optimizeQuery); err != nil {
		slog.Warn("failed to optimize table, indexes may not apply immediately", "tableName", projectID, "error", err)
	}

	return nil
}

func GetProjectsID() (error) {
	rows, err := database.Query("SELECT id FROM projects")
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("scan error: %w", err)
		}
		ProjectIDs = append(ProjectIDs, id)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}

	return nil
}

func StartUP(conn clickhouse.Conn) {
	for i := 0; i < len(ProjectIDs); i++ {
		createTable(conn,ProjectIDs[i])
	}
}

func printProject() {
	for i := 0; i < len(ProjectIDs); i++ {
		log.Printf("table inside of cockroachdb ")
	}
}

