package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
)

func FakeAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simulated logged-in user
		user := &common.User{
			Username: "test_user",
			Password: "test_pass",
			Projects: []common.Project{
				{
					ID:             "p1",
					ApiKey:         "api_key_1",
					Name:           "Order System",
					SearchableKeys: []string{"user_id", "role", "device"},
					OtherKeys:      []string{"ip1", "browser1"}},
				{
					ID:             "p2",
					ApiKey:         "api_key_2",
					Name:           "Analytics Platform",
					SearchableKeys: []string{"age", "height"},
					OtherKeys:      []string{"ip1", "browser1"}},
				{
					ID:             "p3",
					ApiKey:         "api_key_3",
					Name:           "Marketing App",
					SearchableKeys: []string{"tvMarketing", "radioMarketing"},
					OtherKeys:      []string{"ip1", "browser1"}},
			},
		}
		// Store in context
		c.Set("user", user)
		c.Next()
	}
}
