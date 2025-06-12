package common

type GroupedEventsResponse struct {
	Page    int                 `json:"page"`
	HasNext bool                `json:"has_next"`
	Groups  []EventGroupSummary `json:"groups"`
}

type EventGroupSummary struct {
	Name     string // event name (clickable)
	LastSeen string // latest timestamp (among matching events)
	Total    int    // how many matched
}

// GuiEvent TODO: this is temporary
type GuiEvent struct {
	ID             string            // unique ID (e.g., UUID)
	Name           string            // event name
	Timestamp      string            // when it happened
	InsertedAt     string            // when it was saved (optional)
	SearchableKeys map[string]string // searchable key-value fields
	OtherKeys      map[string]string // non-searchable key-value fields
}

type EventDetailResponse struct {
	GuiEvent    GuiEvent          `json:"event"`
	HasPrev     bool              `json:"has_prev"`
	HasNext     bool              `json:"has_next"`
	Filters     map[string]string `json:"filters"`
	Index       int               `json:"index"`
	ProjectID   string            `json:"project_id"`
	ProjectName string            `json:"project_name"`
}
