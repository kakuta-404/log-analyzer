package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	"net/http"
)

func ShowProjectsPage(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.String(401, "Unauthorized")
		return
	}

	usr := user.(*common.User)

	c.HTML(http.StatusOK, "projects_list.gohtml", gin.H{
		"Projects": usr.Projects,
	})
}
