package cassandra

import (
	"context"
	"log/slog"
	"time"

	"github.com/kakuta-404/log-analyzer/common"
)

func GetEventGroupSummaries(projectID string) ([]common.EventGroupSummary, error) {
	slog.Info("starting GetEventGroupSummaries", "projectID", projectID)

	query := `
		SELECT name, MAX(event_timestamp) as last_seen, COUNT(*) as total
		FROM log_analyzer.events
		WHERE project_id = ?
		GROUP BY name
		ALLOW FILTERING
	`
	slog.Debug("executing query", "query", query)

	iter := Session.Query(query, projectID).WithContext(context.Background()).Iter()
	slog.Debug("query iterator created")

	var results []common.EventGroupSummary
	var name string
	var lastSeen time.Time
	var total int

	recordCount := 0
	for iter.Scan(&name, &lastSeen, &total) {
		slog.Debug("scanning row",
			"name", name,
			"lastSeen", lastSeen,
			"total", total)

		results = append(results, common.EventGroupSummary{
			Name:     name,
			LastSeen: lastSeen.Format("2006-01-02 15:04"),
			Total:    total,
		})
		recordCount++
	}

	if err := iter.Close(); err != nil {
		slog.Error("error closing iterator",
			"error", err,
			"projectID", projectID)
		return nil, err
	}

	slog.Info("completed GetEventGroupSummaries",
		"projectID", projectID,
		"recordCount", recordCount,
		"resultsLength", len(results))

	return results, nil
}
