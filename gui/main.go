package main

import (
	"GUI/internal/fake"
	"GUI/internal/handlers"
	"GUI/internal/middleware"
	"GUI/ui/templates"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	//TODO replace with real data
	fake.LoadFakeDataOnce()

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

	// TODO: replace fakeauth with real auth middleware
	router.Use(middleware.FakeAuth())

	// Register routes
	router.GET("/", handlers.ShowHomePage)
	router.GET("/login", handlers.ShowLoginPage)
	router.GET("/signup", handlers.ShowSignupPage)
	router.GET("/projects", handlers.ShowProjectsPage)
	router.GET("/events", handlers.ShowEventList)
	router.GET("/search", handlers.ShowSearchPage)
	router.POST("/search", handlers.ShowSearchPage)
	router.GET("/search/detail", handlers.ShowEventDetail)

	// Use port 8082
	log.Fatal(router.Run(":8082"))
}
