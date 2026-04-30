package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/diyas/fintrack/config"
	v1 "github.com/diyas/fintrack/internal/delivery/rest/v1"
	"github.com/diyas/fintrack/internal/service"
	"github.com/diyas/fintrack/pkg/logger"
)

type Handler struct {
	services *service.Service
	logger   *logger.Logger
}

func NewHandler(services *service.Service, appLogger *logger.Logger) *Handler {
	return &Handler{
		services: services,
		logger:   appLogger,
	}
}

func (h *Handler) Init(cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Add custom middleware
	router.Use(
		logger.RequestIDMiddleware(),
		logger.LoggingMiddleware(h.logger),
		logger.RecoveryMiddleware(h.logger),
		logger.ContextMiddleware(h.logger),
	)

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	h.initAPI(router, cfg)

	return router
}

func (h *Handler) initAPI(router *gin.Engine, cfg *config.Config) {
	handlerV1 := v1.NewHandler(h.services, h.logger, cfg)
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
