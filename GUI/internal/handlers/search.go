package handlers

import (
	"GUI/internal/fake"
	"GUI/internal/logic"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ShowSearchPage(c *gin.Context) {
	user, ok := getAuthenticatedUser(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	projectName, searchKeys, ok := validateProjectAccess(c, user, projectID)
	if !ok {
		return
	}

	filters := map[string]string{}
	if c.Request.Method == http.MethodPost {
		for _, key := range searchKeys {
			val := c.PostForm(key)
			if val != "" {
				filters[key] = val
			}
		}
	}

	// Get and filter events
	all := fake.ProjectEvents[projectID]
	filtered := logic.FilterEvents(all, filters)
	grouped := logic.GroupEventsByName(filtered)

	c.HTML(http.StatusOK, "search.gohtml", gin.H{
		"ProjectID":   projectID,
		"ProjectName": projectName,
		"Filters":     filters,
		"SearchKeys":  searchKeys,
		"Groups":      grouped,
	})
}
