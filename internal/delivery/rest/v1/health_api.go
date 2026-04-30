package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initHealthRoutes(router *gin.RouterGroup) {
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"request_id": c.GetString("request_id"),
		})
	})

	router.GET("/readyz", func(c *gin.Context) {
		if err := h.services.Health.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":     "not ready",
				"error":      err.Error(),
				"request_id": c.GetString("request_id"),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":     "ready",
			"request_id": c.GetString("request_id"),
		})
	})
}
