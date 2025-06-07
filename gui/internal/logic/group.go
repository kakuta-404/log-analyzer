package logic

import (
	"GUI/internal/models"
)

func GroupEventsByName(events []models.Event) []models.EventGroupSummary {
	//TODO replace with fast db group by method
	grouped := map[string]models.EventGroupSummary{}

	for _, e := range events {
		g := grouped[e.Name]
		g.Name = e.Name
		g.Total++
		// compare timestamp
		if e.Timestamp > g.LastSeen {
			g.LastSeen = e.Timestamp
		}
		grouped[e.Name] = g
	}

	var result []models.EventGroupSummary
	for _, v := range grouped {
		result = append(result, v)
	}
	return result
}
