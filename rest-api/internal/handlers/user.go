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
				Name:           "Order System",
				ApiKey:         "key1",
				SearchableKeys: []string{"level", "message", "service", "registered"},
				OtherKeys:      []string{"ip", "browser"},
			},
			{
				ID:             "p2",
				Name:           "Analytics Platform",
				ApiKey:         "key2",
				SearchableKeys: []string{"user_id"},
				OtherKeys:      []string{"region"},
			},
		},
	}
	c.JSON(http.StatusOK, user)
}
