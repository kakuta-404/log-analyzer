package main

import (
	"github.com/gin-gonic/gin"
)

type LogPayload struct {
	ProjectID string            `json:"project_id"`
	APIKey    string            `json:"api_key"`
	Name      string            `json:"name"`
	Timestamp int64             `json:"timestamp"`
	Data      map[string]string `json:"data"`
}

func main() {
	r := gin.Default()

	r.POST("/logs", func(c *gin.Context) {
		var payload LogPayload
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		// TODO: Validate API key and send to Kafka
		c.JSON(200, gin.H{"status": "received"})
	})

	r.Run(":8080")
}
