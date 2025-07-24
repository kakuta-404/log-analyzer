package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
)

func GetCurrentUser(c *gin.Context) {
	//TODO : Replace with actual authentication logic
	user := common.User{
		Username: "test-user",
		Password: "test-password",
		Projects: []common.Project{
			{
				ID:             "test-project",
				Name:           "test-project",
				ApiKey:         "test-key",
				SearchableKeys: []string{"level", "message", "service", "registered"},
				OtherKeys:      []string{"ip", "browser"},
			},
		},
	}
	c.JSON(http.StatusOK, user)
}
