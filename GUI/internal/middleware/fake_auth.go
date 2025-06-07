package middleware

import (
	"GUI/internal/models"
	"github.com/gin-gonic/gin"
)

func FakeAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simulated logged-in user
		user := &models.User{
			ID:       "u1",
			Username: "test_user",
			Projects: []models.Project{
				{
					ID:             "p1",
					Name:           "Order System",
					SearchableKeys: []string{"user_id", "role", "device"},
					OtherKeys:      []string{"ip1", "browser1"}},
				{
					ID:             "p2",
					Name:           "Analytics Platform",
					SearchableKeys: []string{"age", "height"},
					OtherKeys:      []string{"ip1", "browser1"}},
				{
					ID:             "p3",
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
