package handlers

import (
	"GUI/internal/fake"
	"GUI/internal/logic"
	"GUI/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ShowSearchPage(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.String(http.StatusBadRequest, "Missing project_id")
		return
	}

	// Fake auth
	userVal, exists := c.Get("user")
	if !exists {
		c.String(http.StatusUnauthorized, "unauthorized")
		return
	}
	user := userVal.(*models.User)

	// Check project access
	var (
		projectName string
		searchKeys  []string
		found       bool
	)

	for _, p := range user.Projects {
		if p.ID == projectID {
			projectName = p.Name
			searchKeys = p.SearchableKeys // ✅ get schema here
			found = true
			break
		}
	}

	if !found {
		c.String(http.StatusForbidden, "no access")
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
