package v1

import (
	"net/http"

	"github.com/diyas/fintrack/internal/delivery/dto"
	"github.com/diyas/fintrack/internal/delivery/middleware"
	"github.com/diyas/fintrack/pkg/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initExchangeRateRoutes(router *gin.RouterGroup) {
	router.GET("/exchange-rates", middleware.Auth(h.services.Auth.TokenManager()), h.getExchangeRate)
}

func (h *Handler) getExchangeRate(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "from and to query parameters are required", nil)
		return
	}
	er, err := h.services.Currency.GetRateWithMeta(c.Request.Context(), from, to)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ExchangeRateResponseFrom(er))
}
