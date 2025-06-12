package logic

import (
	"github.com/kakuta-404/log-analyzer/common"
)

func GroupEventsByName(events []common.GuiEvent) []common.EventGroupSummary {
	//TODO replace with fast db group by method
	grouped := map[string]common.EventGroupSummary{}

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

	var result []common.EventGroupSummary
	for _, v := range grouped {
		result = append(result, v)
	}
	return result
}
