package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/handlers"
)

func main() {
	router := gin.Default()

	router.GET("/api/user", handlers.GetCurrentUser)

	router.GET("/projects/:id/events/:name", func(c *gin.Context) {
		// TODO: Implement event details
		c.JSON(200, gin.H{"status": "not implemented"})
	})

	router.Run(common.RESTAPIPort)
}
