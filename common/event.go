package common

import "time"

type Event struct {
	Name           string
	ProjectID      string
	EventTimestamp time.Time
	Log            map[string]string
}
