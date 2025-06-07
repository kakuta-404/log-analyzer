package main

import (
	"GUI/internal/handlers"
	"GUI/ui/templates"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Configure trusted proxies
	router.SetTrustedProxies([]string{"127.0.0.1"})

	// Initialize templates
	htmlRender, err := templates.Instance()
	if err != nil {
		log.Fatal(err)
	}

	// Set HTML renderer
	router.HTMLRender = htmlRender

	// Serve static files
	router.Static("/static", "./ui/static")

	// Register routes
	router.GET("/login", handlers.Login)
	router.GET("/events", handlers.EventList)
	router.GET("/search", handlers.SearchPage)

	// Use port 8082
	log.Fatal(router.Run(":8082"))
}
