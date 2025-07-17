package clickhouse

import (
	"context"
	_ "encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/kakuta-404/log-analyzer/common"
)

func GetFilteredGroupedEvents(projectID string, filters map[string]string) ([]common.EventGroupSummary, error) {
	slog.Info("Starting GetFilteredGroupedEvents", "projectID", projectID, "filters", filters)

	var conditions []string
	params := []any{projectID}

	// Build WHERE clause from filters
	for key, val := range filters {
		conditions = append(conditions, fmt.Sprintf("log_data['%s'] = ?", key))
		params = append(params, val)
	}

	where := "WHERE project_id = ?"
	if len(conditions) > 0 {
		where += " AND " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT 
			name, 
			MAX(timestamp) AS last_seen, 
			COUNT(*) AS total
		FROM events
		%s
		GROUP BY name
		ORDER BY last_seen DESC
		LIMIT 100
	`, where)

	slog.Info("Executing query",
		"query", query,
		"params", params)

	rows, err := Conn.Query(context.Background(), query, params...)
	if err != nil {
		slog.Error("Query execution failed",
			"error", err,
			"projectID", projectID)
		return nil, err
	}
	defer rows.Close()

	var results []common.EventGroupSummary
	for rows.Next() {
		var name string
		var lastSeen time.Time
		var total uint64

		if err := rows.Scan(&name, &lastSeen, &total); err != nil {
			return nil, err
		}

		results = append(results, common.EventGroupSummary{
			Name:     name,
			LastSeen: lastSeen.Format("2006-01-02 15:04"), // same as UI format
			Total:    int(total),
		})
	}

	slog.Info("Completed GetFilteredGroupedEvents",
		"projectID", projectID,
		"resultCount", len(results))

	return results, nil
}

func GetFilteredEventDetail(projectID, eventName string, filters map[string]string, index int) (*common.GuiEvent, bool, bool, error) {
	slog.Info("Starting GetFilteredEventDetail",
		"projectID", projectID,
		"eventName", eventName,
		"filters", filters,
		"index", index)

	var conditions []string
	params := []any{projectID, eventName}

	for key, val := range filters {
		conditions = append(conditions, fmt.Sprintf("log_data['%s'] = ?", key))
		params = append(params, val)
	}

	where := "WHERE project_id = ? AND name = ?"
	if len(conditions) > 0 {
		where += " AND " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT 
			project_id,
			timestamp,
			log_data
		FROM events
		%s
		ORDER BY timestamp, project_id
		LIMIT 3 OFFSET ?
	`, where)

	params = append(params, index-1)

	slog.Debug("Executing query",
		"query", query,
		"params", params)

	rows, err := Conn.Query(context.Background(), query, params...)
	if err != nil {
		slog.Error("Query execution failed",
			"error", err,
			"projectID", projectID,
			"eventName", eventName)
		return nil, false, false, err
	}
	defer rows.Close()

	var events []common.GuiEvent
	for rows.Next() {
		var projectId string
		var ts time.Time
		var logData map[string]string

		if err := rows.Scan(&projectId, &ts, &logData); err != nil {
			return nil, false, false, err
		}

		events = append(events, common.GuiEvent{
			ID:             projectId,
			Name:           eventName,
			Timestamp:      ts.Format("2006-01-02 15:04"),
			SearchableKeys: logData,
			OtherKeys:      map[string]string{},
		})
	}

	if index == 0 && len(events) == 0 {
		return nil, false, false, nil
	}

	var event *common.GuiEvent
	var hasPrev, hasNext bool

	if index == 0 {
		if len(events) > 0 {
			event = &events[0]
			hasNext = len(events) > 1
		}
	} else {
		if len(events) == 3 {
			event = &events[1]
			hasPrev = true
			hasNext = true
		} else if len(events) == 2 {
			event = &events[1]
			hasPrev = true
		} else if len(events) == 1 {
			event = &events[0]
			hasPrev = true
		}
	}

	slog.Info("Completed GetFilteredEventDetail",
		"projectID", projectID,
		"eventName", eventName,
		"foundEvent", event != nil,
		"hasPrev", hasPrev,
		"hasNext", hasNext)

	return event, hasPrev, hasNext, nil
}
