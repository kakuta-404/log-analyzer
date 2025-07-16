package cassandra

import (
	"context"
	"github.com/kakuta-404/log-analyzer/common"
	"time"
)

func GetEventGroupSummaries(projectID string) ([]common.EventGroupSummary, error) {
	query := `
		SELECT name, MAX(timestamp) as last_seen, COUNT(*) as total
		FROM events_by_name
		WHERE project_id = ?
		GROUP BY name
		ALLOW FILTERING
	`

	iter := Session.Query(query, projectID).WithContext(context.Background()).Iter()

	var results []common.EventGroupSummary
	var name string
	var lastSeen time.Time
	var total int

	for iter.Scan(&name, &lastSeen, &total) {
		results = append(results, common.EventGroupSummary{
			Name:     name,
			LastSeen: lastSeen.Format("2006-01-02 15:04"),
			Total:    total,
		})
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}

	return results, nil
}
