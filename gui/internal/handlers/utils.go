package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	"net/http"
)

// getAuthenticatedUser retrieves authenticated user from context
func getAuthenticatedUser(c *gin.Context) (*common.User, bool) {
	userVal, exists := c.Get("user")
	if !exists {
		c.String(http.StatusUnauthorized, "unauthorized")
		return nil, false
	}
	return userVal.(*common.User), true
}

// validateProjectAccess checks if user has access to project and returns project details
func validateProjectAccess(c *gin.Context, user *common.User, projectID string) (projectName string, searchableKeys []string, ok bool) {
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
