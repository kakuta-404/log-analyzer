package handler

import (
	"context"
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
	var sub common.Submission
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.auth.ValidateAPIKey(sub.ProjectID, sub.APIKey); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
		return
	}

	event := common.Event{
		Name:           sub.Name,
		ProjectID:      sub.ProjectID,
		EventTimestamp: sub.Timestamp,
		Log:            sub.PayLoad,
	}

	if err := h.producer.SendEvent(c.Request.Context(), &event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process log"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "log submitted successfully"})
}
