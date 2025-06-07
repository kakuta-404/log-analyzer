package handlers

import (
	"GUI/internal/fake"
	"GUI/internal/logic"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

const GroupsPerPage = 10

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

	// TODO: get real events
	events := fake.ProjectEvents[projectID]
	grouped := logic.GroupEventsByName(events)

	// Pagination
	start := (page - 1) * GroupsPerPage
	end := start + GroupsPerPage
	if start > len(grouped) {
		start = len(grouped)
	}
	if end > len(grouped) {
		end = len(grouped)
	}
	paged := grouped[start:end]

	c.HTML(http.StatusOK, "event_list.gohtml", gin.H{
		"ProjectID":   projectID,
		"ProjectName": projectName,
		"Page":        page,
		"HasNext":     end < len(grouped),
		"Events":      paged,
	})
}
