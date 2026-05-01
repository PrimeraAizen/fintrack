package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/diyas/fintrack/config"
	"github.com/diyas/fintrack/internal/service"
	"github.com/diyas/fintrack/pkg/logger"
)

type Handler struct {
	services *service.Service
	logger   *logger.Logger
	cfg      *config.Config
}

func NewHandler(services *service.Service, appLogger *logger.Logger, cfg *config.Config) *Handler {
	return &Handler{
		services: services,
		logger:   appLogger,
		cfg:      cfg,
	}
}

func (h *Handler) Init(api *gin.RouterGroup) {
	v1 := api.Group("/v1")
	h.initHealthRoutes(v1)
	h.initAuthRoutes(v1)
	h.initAccountRoutes(v1)
	h.initCategoryRoutes(v1)
	h.initTransactionRoutes(v1)
	h.initBudgetRoutes(v1)
	h.initTransferRoutes(v1)
	h.initSavingGoalRoutes(v1)
	h.initReportRoutes(v1)
	h.initCSVRoutes(v1)
	h.initExchangeRateRoutes(v1)
}
