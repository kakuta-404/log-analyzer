package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	"net/http"
)

func ShowEventDetail(c *gin.Context) {
	user, ok := getAuthenticatedUser(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	name := c.Query("name")
	indexStr := c.DefaultQuery("index", "0")

	projectName, searchableKeys, ok := validateProjectAccess(c, user, projectID)
	if !ok {
		return
	}

	// Filters
	filters := map[string]string{}
	q := fmt.Sprintf("%s/projects/%s/events/%s/detail?index=%s", common.RESTAPIBaseURL, projectID, name, indexStr)
	for _, key := range searchableKeys {
		if val := c.Query(key); val != "" {
			filters[key] = val
			q += fmt.Sprintf("&%s=%s", key, val)
		}
	}

	// Request
	resp, err := http.Get(q)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.String(http.StatusBadGateway, "Failed to fetch event detail")
		return
	}
	defer resp.Body.Close()

	var result common.EventDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.String(http.StatusInternalServerError, "Invalid response")
		return
	}

	// Render
	c.HTML(http.StatusOK, "event_detail.gohtml", map[string]any{
		"ProjectID":   result.ProjectID,
		"ProjectName": projectName,
		"Event":       result.GuiEvent,
		"Filters":     result.Filters,
		"Index":       result.Index,
		"HasPrev":     result.HasPrev,
		"HasNext":     result.HasNext,
	})
}
