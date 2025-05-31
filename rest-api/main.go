package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/projects/:id/events", func(c *gin.Context) {
		// TODO: Implement event listing
		c.JSON(200, gin.H{"status": "not implemented"})
	})

	r.GET("/projects/:id/events/:name", func(c *gin.Context) {
		// TODO: Implement event details
		c.JSON(200, gin.H{"status": "not implemented"})
	})

	r.Run(":8081")
}
