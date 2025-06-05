package handlers

import (
	"GUI/internal/fake"
	"GUI/internal/logic"
	"GUI/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

const GroupsPerPage = 10

func ShowEventList(c *gin.Context) {
	projectID := c.Query("project_id")
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	userVal, exists := c.Get("user")
	if !exists {
		c.String(http.StatusUnauthorized, "unauthorized")
		return
	}
	user := userVal.(*models.User)

	// Check access
	var projectName string
	allowed := false
	for _, p := range user.Projects {
		if p.ID == projectID {
			projectName = p.Name
			allowed = true
			break
		}
	}
	if !allowed {
		c.String(http.StatusForbidden, "no access to project")
		return
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
