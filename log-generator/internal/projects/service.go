package projects

import (
	"database/sql"
	"encoding/json"
	"log/slog"

	_ "github.com/lib/pq"
)

type Project struct {
	ID             string
	APIKey         string
	SearchableKeys []string
}

type Service struct {
	db *sql.DB
}

func NewService(cfg struct{ DSN string }) (*Service, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create projects table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id STRING PRIMARY KEY,
			api_key STRING NOT NULL,
			searchable_keys JSONB DEFAULT '[]'::JSONB
		)
	`)
	if err != nil {
		slog.Error("failed to create projects table", "error", err)
		return nil, err
	}

	// Insert test data if table is empty
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count); err != nil {
		return nil, err
	}

	if count == 0 {
		slog.Info("inserting test project data")
		keys, _ := json.Marshal([]string{"level", "message", "service", "registered"})
		_, err = db.Exec(`
			INSERT INTO projects (id, api_key, searchable_keys) 
			VALUES ('test_project', 'test-key', $1)
		`, keys)
		if err != nil {
			slog.Error("failed to insert test data", "error", err)
			return nil, err
		}
	}

	return &Service{db: db}, nil
}

func (s *Service) GetProjects() ([]Project, error) {
	rows, err := s.db.Query("SELECT id, api_key, searchable_keys FROM projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		var keys []byte
		if err := rows.Scan(&p.ID, &p.APIKey, &keys); err != nil {
			slog.Error("failed to scan project row", "error", err)
			continue
		}

		if err := json.Unmarshal(keys, &p.SearchableKeys); err != nil {
			slog.Error("failed to parse searchable keys", "error", err)
			continue
		}

		projects = append(projects, p)
	}

	slog.Info("fetched projects", "count", len(projects))
	return projects, nil
}
