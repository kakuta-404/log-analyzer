package common

import "time"

type Event struct {
	Name           string
	ProjectID      string
	EventTimestamp time.Time
	Log            map[string]string
}

type Submission struct {
	ProjectID string `json:"project_id"`
	APIKey    string `json:"api_key"`
	Name      string `json:"name"`
	Timestamp int32 `json:"timestamp"`
	PayLoad map[string]string `json:"payload"`
}
