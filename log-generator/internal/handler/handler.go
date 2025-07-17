package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/log-generator/internal/generator"
)

type Handler struct {
	gen    *generator.Generator
	server *http.Server
}

func New(gen *generator.Generator) *Handler {
	return &Handler{gen: gen}
}

func (h *Handler) Start(port string) {
	r := gin.Default()
	r.POST("/send-now", h.sendNow)
	h.server = &http.Server{
		Addr:    port,
		Handler: r,
	}
	if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func (h *Handler) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func (h *Handler) sendNow(c *gin.Context) {
	if err := h.gen.GenerateAndSend(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send log"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "log sent manually"})
}
