package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/storage/clickhouse"
)

func GetEventDetail(c *gin.Context) {
	projectID := c.Param("id")
	eventName := c.Param("name")
	slog.Info("Getting event detail",
		"projectID", projectID,
		"eventName", eventName)

	indexStr := c.DefaultQuery("index", "0")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 {
		slog.Warn("Invalid index provided", "index", indexStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid index"})
		return
	}

	filters := map[string]string{}
	for key, vals := range c.Request.URL.Query() {
		if key != "index" {
			filters[key] = vals[0]
		}
	}
	slog.Info("Applied filters", "filters", filters)

	event, hasPrev, hasNext, err := clickhouse.GetFilteredEventDetail(projectID, eventName, filters, index)
	if err != nil {
		slog.Error("Failed to fetch event detail",
			"error", err,
			"projectID", projectID,
			"eventName", eventName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if event == nil {
		slog.Info("No event found", "index", index)
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	slog.Info("Retrieved event detail",
		"index", index,
		"hasPrev", hasPrev,
		"hasNext", hasNext)

	resp := common.EventDetailResponse{
		GuiEvent:    *event,
		Index:       index,
		HasPrev:     hasPrev,
		HasNext:     hasNext,
		Filters:     filters,
		ProjectID:   projectID,
		ProjectName: "(placeholder)",
	}

	c.JSON(http.StatusOK, resp)
}
