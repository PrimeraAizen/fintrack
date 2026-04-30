package v1

import (
	"net/http"

	"github.com/diyas/fintrack/internal/delivery/dto"
	"github.com/diyas/fintrack/internal/delivery/middleware"
	"github.com/diyas/fintrack/pkg/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initTransferRoutes(router *gin.RouterGroup) {
	transfers := router.Group("/transfers", middleware.Auth(h.services.Auth.TokenManager()))
	{
		transfers.POST("", h.createTransfer)
		transfers.GET("", h.listTransfers)
	}
}

func (h *Handler) createTransfer(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	var req dto.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	t, err := h.services.Transfer.Execute(c.Request.Context(), userID, req.ToInput())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, dto.TransferResponseFrom(t))
}

func (h *Handler) listTransfers(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	transfers, err := h.services.Transfer.List(c.Request.Context(), userID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.TransferListResponseFrom(transfers))
}
