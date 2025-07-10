package logic

import (
	"github.com/kakuta-404/log-analyzer/common"
)

func FilterEvents(events []common.GuiEvent, filters map[string]string) []common.GuiEvent {
	//TODO replace with fast db filter method
	if len(filters) == 0 {
		return events
	}
	var result []common.GuiEvent
	for _, e := range events {
		match := true
		for k, v := range filters {
			if eVal, ok := e.SearchableKeys[k]; !ok || eVal != v {
				match = false
				break
			}
		}
		if match {
			result = append(result, e)
		}
	}
	return result
}
