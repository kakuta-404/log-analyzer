package handler

import (
	"context"
	"log"
	"net/http"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/kakuta-404/log-analyzer/log-drain/internal/auth"
	"github.com/kakuta-404/log-analyzer/log-drain/internal/kafka"
)

type Handler struct {
	auth     *auth.Service
	producer *kafka.Producer
	server   *http.Server
}

func New(auth *auth.Service, producer *kafka.Producer) *Handler {
	return &Handler{
		auth:     auth,
		producer: producer,
	}
}

func (h *Handler) Start(port string) {
	router := gin.Default()
	router.POST("/logs", h.handleLogs)

	h.server = &http.Server{
		Addr:    port,
		Handler: router,
	}

	if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
	}
}

func (h *Handler) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func (h *Handler) handleLogs(c *gin.Context) {
	slog.Info("Received request to /logs", "remote_addr", c.ClientIP()) 

	var sub common.Submission
	
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.auth.ValidateAPIKey(sub.ProjectID, sub.APIKey); err != nil {
		slog.Error("Auth validation failed", "error", err, "project_id", sub.ProjectID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
		return
	}

	event := common.Event{
		Name:           sub.Name,
		ProjectID:      sub.ProjectID,
		EventTimestamp: sub.Timestamp,
		Log:            sub.PayLoad,
	}

	log.Println("lionel messi")

	if err := h.producer.SendEvent(c.Request.Context(), &event); err != nil {
		slog.Error("Failed to send event to Kafka", "error", err, "event", event)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process log"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "log submitted successfully"})
}
