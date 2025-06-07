package logic

import "GUI/internal/models"

func FilterEvents(events []models.Event, filters map[string]string) []models.Event {
	//TODO replace with fast db filter method
	if len(filters) == 0 {
		return events
	}
	var result []models.Event
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
