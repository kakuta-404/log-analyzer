package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/storage/cassandra"
)

const GroupsPerPage = 10

func GetGroupedEvents(c *gin.Context) {
	projectID := c.Param("id")
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	slog.Info("Fetching grouped events",
		"projectID", projectID,
		"page", page)

	// Fetch grouped summaries from Cassandra
	grouped, err := cassandra.GetEventGroupSummaries(projectID)
	if err != nil {
		slog.Error("Failed to fetch event summaries",
			"error", err,
			"projectID", projectID)
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

	slog.Info("Retrieved grouped events",
		"projectID", projectID,
		"page", page,
		"totalGroups", len(grouped),
		"returnedGroups", len(grouped[start:end]))

	c.JSON(http.StatusOK, gin.H{
		"project_id": projectID,
		"page":       page,
		"has_next":   end < len(grouped),
		"groups":     grouped[start:end],
	})
}
