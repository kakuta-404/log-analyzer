package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/handlers"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/storage/cassandra"
	"github.com/kakuta-404/log-analyzer/rest-api/internal/storage/clickhouse"
)

func init() {
	// Configure slog with JSON handler
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})
	slog.SetDefault(slog.New(handler))
}

func main() {
	slog.Info("Starting log-analyzer REST API...")

	// Initialize Cassandra
	slog.Info("Connecting to Cassandra...")
	err := cassandra.Init("cassandra:9042")
	if err != nil {
		slog.Error("Failed to connect to Cassandra", "error", err)
		os.Exit(1)
	}
	slog.Info("Successfully connected to Cassandra")

	// Initialize ClickHouse
	slog.Info("Connecting to ClickHouse...")
	err = clickhouse.Init("clickhouse:9000")
	if err != nil {
		slog.Error("Failed to connect to ClickHouse", "error", err)
		os.Exit(1)
	}
	slog.Info("Successfully connected to ClickHouse")

	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		slog.Info("Health check ping received")
		c.String(http.StatusOK, "pong")
	})

	api := router.Group("/api")
	{
		api.GET("/user", handlers.GetCurrentUser)
	}
	router.GET("/projects/:id/events", handlers.GetGroupedEvents)
	router.GET("/projects/:id/search", handlers.SearchGroupedEvents)
	router.GET("/projects/:id/events/:name/detail", handlers.GetEventDetail)

	slog.Info("Starting REST API server", "port", common.RESTAPIPort)
	if err := router.Run(common.RESTAPIPort); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
