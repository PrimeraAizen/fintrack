package v1

import (
	"net/http"

	"github.com/diyas/fintrack/internal/delivery/dto"
	"github.com/diyas/fintrack/internal/delivery/middleware"
	"github.com/diyas/fintrack/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) initBudgetRoutes(router *gin.RouterGroup) {
	budgets := router.Group("/budgets", middleware.Auth(h.services.Auth.TokenManager()))
	{
		budgets.POST("", h.createBudget)
		budgets.GET("", h.listBudgets)
		budgets.PATCH("/:id", h.updateBudget)
		budgets.DELETE("/:id", h.deleteBudget)
	}
}

func (h *Handler) createBudget(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	var req dto.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	b, err := h.services.Budget.Create(c.Request.Context(), userID, req.ToInput())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, dto.BudgetResponseFrom(b))
}

func (h *Handler) listBudgets(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	views, err := h.services.Budget.List(c.Request.Context(), userID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.BudgetViewResponseFrom(views))
}

func (h *Handler) updateBudget(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid id", nil)
		return
	}
	var req dto.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	b, err := h.services.Budget.Update(c.Request.Context(), userID, id, req.ToInput())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.BudgetResponseFrom(b))
}

func (h *Handler) deleteBudget(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid id", nil)
		return
	}
	if err := h.services.Budget.Delete(c.Request.Context(), userID, id); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"deleted": true})
}
