package handlers

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

func SearchPage(c *gin.Context) {
	c.HTML(http.StatusOK, "search.gohtml", nil)
	slog.Info("Search Page rendered")
}
