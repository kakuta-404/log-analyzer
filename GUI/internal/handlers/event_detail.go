package handlers

import (
	"GUI/internal/fake"
	"GUI/internal/logic"
	"GUI/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func ShowEventDetail(c *gin.Context) {
	user, ok := getAuthenticatedUser(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	projectName, searchableKeys, ok := validateProjectAccess(c, user, projectID)
	if !ok {
		return
	}

	name := c.Query("name")
	indexStr := c.DefaultQuery("index", "0")
	index, _ := strconv.Atoi(indexStr)

	// Parse filters
	filters := map[string]string{}
	for _, key := range searchableKeys {
		if val := c.Query(key); val != "" {
			filters[key] = val
		}
	}

	// Get matching events
	all := fake.ProjectEvents[projectID]
	filtered := logic.FilterEvents(all, filters)

	// Only keep matching name
	var matching []models.Event
	for _, e := range filtered {
		if e.Name == name {
			matching = append(matching, e)
		}
	}

	if index < 0 || index >= len(matching) {
		c.String(http.StatusNotFound, "event not found")
		return
	}

	event := matching[index]

	data := map[string]any{
		"ProjectID":   projectID,
		"ProjectName": projectName,
		"Event":       event,
		"Filters":     filters,
		"Index":       index,
		"HasPrev":     index > 0,
		"HasNext":     index+1 < len(matching),
	}

	c.HTML(http.StatusOK, "event_detail.gohtml", data)

}
