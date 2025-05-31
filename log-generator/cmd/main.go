package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/logs", func(c *gin.Context) {
		// TODO: Implement log generation logic
		c.JSON(http.StatusOK, gin.H{
			"message": "Log received",
		})
	})

	r.Run(":8081")
}
