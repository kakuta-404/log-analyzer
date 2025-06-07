package common

import "time"

type User struct {
	Username string
	Password string
	Projects []Project
}
type Project struct {
	ID             string
	ApiKey         string
	Name           string
	SearchableKeys []string
	OtherKeys      []string
}

type Event struct {
	Name           string
	ProjectID      string
	EventTimestamp time.Time
	Log            map[string]string
}

type Submission struct {
	ProjectID string            `json:"project_id"`
	APIKey    string            `json:"api_key"`
	Name      string            `json:"name"`
	Timestamp time.Time         `json:"event_timestamp"`
	PayLoad   map[string]string `json:"payload"`
}

