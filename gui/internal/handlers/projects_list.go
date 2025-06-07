package handlers

import (
	"GUI/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ShowProjectsPage(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.String(401, "Unauthorized")
		return
	}

	usr := user.(*models.User)

	c.HTML(http.StatusOK, "projects_list.gohtml", gin.H{
		"Projects": usr.Projects,
	})
}
