package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	"net/http"
	"strconv"
)

var result common.GroupedEventsResponse

func ShowEventList(c *gin.Context) {
	user, ok := getAuthenticatedUser(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	projectName, _, ok := validateProjectAccess(c, user, projectID)
	if !ok {
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	// ============================
	// Call REST API for events
	// ============================
	url := fmt.Sprintf("%s/projects/%s/events?page=%d", common.RESTAPIBaseURL, projectID, page)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.String(http.StatusBadGateway, "Failed to fetch events")
		return
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.String(http.StatusInternalServerError, "Invalid response from server")
		return
	}

	// ============================
	// Render HTML
	// ============================
	c.HTML(http.StatusOK, "event_list.gohtml", gin.H{
		"ProjectID":   projectID,
		"ProjectName": projectName,
		"Page":        result.Page,
		"HasNext":     result.HasNext,
		"Events":      result.Groups,
	})
}
