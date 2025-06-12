package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	"net/http"
)

func GetCurrentUser(c *gin.Context) {
	//TODO : Replace with actual authentication logic
	user := common.User{
		Username: "test-user",
		Password: "test-password",
		Projects: []common.Project{
			{
				ID:             "p1",
				Name:           "Order System",
				ApiKey:         "key1",
				SearchableKeys: []string{"user_id", "role"},
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
