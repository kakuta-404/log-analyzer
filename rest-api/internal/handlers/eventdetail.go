package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/fake"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/logic"
	"net/http"
	"strconv"
)

func GetEventDetail(c *gin.Context) {
	projectID := c.Param("id")
	eventName := c.Param("name")
	indexStr := c.DefaultQuery("index", "0")
	index, _ := strconv.Atoi(indexStr)

	// Get filters
	filters := map[string]string{}
	for key, vals := range c.Request.URL.Query() {
		if key != "index" { // Exclude index
			filters[key] = vals[0]
		}
	}

	all := fake.ProjectEvents[projectID]
	filtered := logic.FilterEvents(all, filters)

	var matching []common.GuiEvent
	for _, e := range filtered {
		if e.Name == eventName {
			matching = append(matching, e)
		}
	}

	if index < 0 || index >= len(matching) {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	resp := common.EventDetailResponse{
		GuiEvent:    matching[index],
		Index:       index,
		HasPrev:     index > 0,
		HasNext:     index+1 < len(matching),
		Filters:     filters,
		ProjectID:   projectID,
		ProjectName: "(placeholder)", // or extract from fake user
	}

	c.JSON(http.StatusOK, resp)
}
