package clickhouse

import (
	"context"
	_ "encoding/json"
	"fmt"
	"github.com/kakuta-404/log-analyzer/common"
	"strings"
	"time"
)

func GetFilteredGroupedEvents(projectID string, filters map[string]string) ([]common.EventGroupSummary, error) {
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

	rows, err := Conn.Query(context.Background(), query, params...)
	if err != nil {
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

	return results, nil
}
