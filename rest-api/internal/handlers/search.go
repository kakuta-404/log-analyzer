package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/storage/clickhouse"
)

func SearchGroupedEvents(c *gin.Context) {
	projectID := c.Param("id")
	slog.Info("Searching grouped events", "projectID", projectID)

	filters := map[string]string{}
	for key, vals := range c.Request.URL.Query() {
		if len(vals) > 0 {
			filters[key] = vals[0]
		}
	}
	slog.Info("Applied filters", "filters", filters)

	grouped, err := clickhouse.GetFilteredGroupedEvents(projectID, filters)
	if err != nil {
		slog.Error("Failed to fetch filtered events", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch filtered events"})
		return
	}

	slog.Info("Retrieved event groups", "count", len(grouped))
	c.JSON(http.StatusOK, common.GroupedEventsResponse{
		Groups: grouped,
	})
}
