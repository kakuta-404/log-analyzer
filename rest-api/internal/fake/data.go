package fake

import (
	"github.com/kakuta-404/log-analyzer/common"
)

var ProjectEvents = map[string][]common.GuiEvent{}

func LoadFakeDataOnce() {
	if len(ProjectEvents) == 0 {
		ProjectEvents["p1"] = GenerateFakeEvents("p1", 78)
		ProjectEvents["p2"] = GenerateFakeEvents("p2", 39)
	}
}
