package models

type User struct {
	ID       string
	Username string
	Projects []Project
}

type Project struct {
	ID             string
	Name           string
	SearchableKeys []string
	OtherKeys      []string
}

type Event struct {
	ID             string            // unique ID (e.g., UUID)
	Name           string            // event name
	Timestamp      string            // when it happened
	InsertedAt     string            // when it was saved (optional)
	SearchableKeys map[string]string // searchable key-value fields
	OtherKeys      map[string]string // non-searchable key-value fields
}

type EventGroupSummary struct {
	Name     string // event name (clickable)
	LastSeen string // latest timestamp (among matching events)
	Total    int    // how many matched
}
