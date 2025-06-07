package fake

import (
	"GUI/internal/models"
)

var ProjectEvents = map[string][]models.Event{}

func LoadFakeDataOnce() {
	if len(ProjectEvents) == 0 {
		ProjectEvents["p1"] = GenerateFakeEvents("p1", 78)
		ProjectEvents["p2"] = GenerateFakeEvents("p2", 39)
	}
}
