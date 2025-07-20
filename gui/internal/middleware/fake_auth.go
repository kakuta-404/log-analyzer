package middleware

import (
	"encoding/json"
	"github.com/kakuta-404/log-analyzer/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

func FakeAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := http.Get(common.RESTAPIBaseURL + "/api/user")
		if err != nil || resp.StatusCode != http.StatusOK {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to retrieve user"})
			return
		}
		defer resp.Body.Close()

		var user common.User
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user response"})
			return
		}

		c.Set("user", &user)
		c.Next()
	}
}
