package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/storage/cassandra"
	"net/http"
	"strconv"
)

const GroupsPerPage = 10

func GetGroupedEvents(c *gin.Context) {
	projectID := c.Param("id")
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	// TODO: auth + get events
	//all := fake.ProjectEvents[projectID] // temp
	//grouped := logic.GroupEventsByName(all)

	// Fetch grouped summaries from Cassandra
	grouped, err := cassandra.GetEventGroupSummaries(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch event summaries"})
		return
	}

	start := (page - 1) * GroupsPerPage
	end := start + GroupsPerPage
	if start > len(grouped) {
		start = len(grouped)
	}
	if end > len(grouped) {
		end = len(grouped)
	}

	c.JSON(http.StatusOK, gin.H{
		"project_id": projectID,
		"page":       page,
		"has_next":   end < len(grouped),
		"groups":     grouped[start:end],
	})
}
