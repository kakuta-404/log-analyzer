package handlers

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

func ShowHomePage(c *gin.Context) {
	c.HTML(http.StatusOK, "home.gohtml", nil)
	slog.Info("Home Page rendered")
}
