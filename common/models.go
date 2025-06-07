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
	Name           string            `json:"name"`
	ProjectID      string            `json:"project_id"`
	EventTimestamp time.Time         `json:"event_timestamp"`
	Log            map[string]string `json:"log"`
}

type Submission struct {
	ProjectID string            `json:"project_id"`
	APIKey    string            `json:"api_key"`
	Name      string            `json:"name"`
	Timestamp time.Time         `json:"event_timestamp"`
	PayLoad   map[string]string `json:"payload"`
}
