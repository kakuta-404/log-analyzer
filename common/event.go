package common

import "time"

type Event struct {
	Name           string            `json:"name"`
	ProjectID      string            `json:"project_id"`
	EventTimestamp time.Time         `json:"event_timestamp"`
	Log            map[string]string `json:"log"`
}

type Submission struct {
	ProjectID string            `json:"project_id"`
	APIKey    string            `json:"api_key"`
	Name      string            `json:"name"`
	Timestamp time.Time             `json:"timestamp"`
	PayLoad   map[string]string `json:"payload"`
}
