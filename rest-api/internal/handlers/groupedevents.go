package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/fake"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/logic"
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
	all := fake.ProjectEvents[projectID] // temp
	grouped := logic.GroupEventsByName(all)

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
