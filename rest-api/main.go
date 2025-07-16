package main

import (
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/fake"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/storage/cassandra"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/storage/clickhouse"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/handlers"
)

func main() {

	//////TODO: replace with real Cassandra and ClickHouse connections
	err := cassandra.Init("localhost:9042") // replace with container if needed
	if err != nil {
		log.Fatal("failed to connect to cassandra:", err)
	}

	err = clickhouse.Init("localhost:9000")
	if err != nil {
		log.Fatal("failed to connect to clickhouse:", err)
	}
	//TODO: replace with real data
	//fake.LoadFakeDataOnce()

	router := gin.Default()

	// Trust only localhost (adjust if needed)
	//if err := router.SetTrustedProxies(nil); err != nil {
	//	log.Fatal("Failed to set trusted proxies:", err)
	//}

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	api := router.Group("/api")
	{
		api.GET("/user", handlers.GetCurrentUser)
	}
	router.GET("/projects/:id/events", handlers.GetGroupedEvents)
	router.GET("/projects/:id/search", handlers.SearchGroupedEvents)
	router.GET("/projects/:id/events/:name/detail", handlers.GetEventDetail)

	log.Println("Starting REST API on", common.RESTAPIPort)
	if err := router.Run(common.RESTAPIPort); err != nil {
		log.Fatal("Failed to start REST API:", err)
	}
}
