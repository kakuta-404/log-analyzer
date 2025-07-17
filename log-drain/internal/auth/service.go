package auth

import (
	"database/sql"
	"errors"
	"sync"

	_ "github.com/lib/pq"
)

type Service struct {
	db    *sql.DB
	cache sync.Map // project_id -> api_key
}

func NewService(cfg struct{ DSN string }) (*Service, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Service{
		db: db,
	}, nil
}

func (s *Service) ValidateAPIKey(projectID, apiKey string) error {
	// Check cache first
	if cachedKey, ok := s.cache.Load(projectID); ok {
		if cachedKey == apiKey {
			return nil
		}
	}

	// Query database
	var dbAPIKey string
	err := s.db.QueryRow("SELECT api_key FROM projects WHERE id = $1", projectID).Scan(&dbAPIKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("project not found")
		}
		return err
	}

	if dbAPIKey != apiKey {
		return errors.New("invalid API key")
	}

	// Update cache
	s.cache.Store(projectID, apiKey)
	return nil
}
