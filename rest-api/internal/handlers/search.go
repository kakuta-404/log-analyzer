package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/fake"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/logic"
	"net/http"
)

func SearchGroupedEvents(c *gin.Context) {
	projectID := c.Param("id")
	//TODO: (Optional) check project exists if needed

	filters := map[string]string{}
	for key, vals := range c.Request.URL.Query() {
		if len(vals) > 0 {
			filters[key] = vals[0]
		}
	}

	// Replace with real data source later
	all := fake.ProjectEvents[projectID]
	filtered := logic.FilterEvents(all, filters)
	grouped := logic.GroupEventsByName(filtered)

	c.JSON(http.StatusOK, common.GroupedEventsResponse{
		Groups: grouped,
	})
}
