package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type EventSummary struct {
	EventName string
	LastSeen  string
	Total     int
}

func EventList(c *gin.Context) {
	// TODO: Call REST API to get event summary list
	sample := []EventSummary{
		{"signup", "2025-05-30 12:00", 34},
		{"login", "2025-05-30 13:20", 123},
	}
	c.HTML(http.StatusOK, "event_list.gohtml", sample)
}
