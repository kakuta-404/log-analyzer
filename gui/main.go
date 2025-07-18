package main

import (
	"GUI/internal/handlers"
	"GUI/internal/middleware"
	"GUI/ui/templates"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
)

func main() {

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	htmlRender, err := templates.Instance()
	if err != nil {
		log.Fatal(err)
	}

	router.HTMLRender = htmlRender

	router.Static("/static", "/app/ui/static")

	router.Use(middleware.Auth())

	router.GET("/", handlers.ShowHomePage)
	router.GET("/login", handlers.ShowLoginPage)
	router.GET("/signup", handlers.ShowSignupPage)
	router.GET("/projects", handlers.ShowProjectsPage)
	router.GET("/events", handlers.ShowEventList)
	router.GET("/search", handlers.ShowSearchPage)
	router.POST("/search", handlers.ShowSearchPage)
	router.GET("/search/detail", handlers.ShowEventDetail)

	log.Fatal(router.Run(common.GUIBaseURL))
}
