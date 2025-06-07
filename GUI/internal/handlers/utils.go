package handlers

import (
	"GUI/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

// getAuthenticatedUser retrieves authenticated user from context
func getAuthenticatedUser(c *gin.Context) (*models.User, bool) {
	userVal, exists := c.Get("user")
	if !exists {
		c.String(http.StatusUnauthorized, "unauthorized")
		return nil, false
	}
	return userVal.(*models.User), true
}

// validateProjectAccess checks if user has access to project and returns project details
func validateProjectAccess(c *gin.Context, user *models.User, projectID string) (projectName string, searchableKeys []string, ok bool) {
	if projectID == "" {
		c.String(http.StatusBadRequest, "Missing project_id")
		return "", nil, false
	}

	for _, p := range user.Projects {
		if p.ID == projectID {
			return p.Name, p.SearchableKeys, true
		}
	}

	c.String(http.StatusForbidden, "no access to project")
	return "", nil, false
}
