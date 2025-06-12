//package main
//
//import (
//	"github.com/gin-gonic/gin"
//	"github.com/kakuta-404/log-analyzer/common"
//	"github.com/kakuta-404/log-analyzer/rest-api/internal/handlers"
//)
//
//func main() {
//	router := gin.Default()
//
//	router.GET("/api/user", handlers.GetCurrentUser)
//
//	router.GET("/projects/:id/events/:name", func(c *gin.Context) {
//		// TODO: Implement event details
//		c.JSON(200, gin.H{"status": "not implemented"})
//	})
//
//	router.Run(common.RESTAPIPort)
//}

package main

import (
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/fake"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/handlers"
)

func main() {
	//TODO: replace with real data
	fake.LoadFakeDataOnce()

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
