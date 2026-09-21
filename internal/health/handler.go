package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Pinger interface {
	PingContext(context.Context) error
}

type Handler struct {
	database Pinger
}

func NewHandler(database Pinger) *Handler {
	return &Handler{database: database}
}

func (handler *Handler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (handler *Handler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := handler.database.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": gin.H{
				"code":    "not_ready",
				"message": "required dependencies are unavailable",
			},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
