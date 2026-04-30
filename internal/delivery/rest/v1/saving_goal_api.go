package v1

import (
	"net/http"

	"github.com/diyas/fintrack/internal/delivery/dto"
	"github.com/diyas/fintrack/internal/delivery/middleware"
	"github.com/diyas/fintrack/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) initSavingGoalRoutes(router *gin.RouterGroup) {
	goals := router.Group("/saving-goals", middleware.Auth(h.services.Auth.TokenManager()))
	{
		goals.POST("", h.createSavingGoal)
		goals.GET("", h.listSavingGoals)
		goals.PUT("/:id", h.updateSavingGoal)
		goals.POST("/:id/contribute", h.contributeSavingGoal)
		goals.DELETE("/:id", h.deleteSavingGoal)
	}
}

func (h *Handler) createSavingGoal(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	var req dto.CreateSavingGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	g, err := h.services.SavingGoal.Create(c.Request.Context(), userID, req.ToInput())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, dto.SavingGoalResponseFrom(g))
}

func (h *Handler) listSavingGoals(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	views, err := h.services.SavingGoal.List(c.Request.Context(), userID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.SavingGoalListResponseFrom(views))
}

func (h *Handler) updateSavingGoal(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid id", nil)
		return
	}
	var req dto.UpdateSavingGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	g, err := h.services.SavingGoal.Update(c.Request.Context(), userID, id, req.ToInput())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.SavingGoalResponseFrom(g))
}

func (h *Handler) contributeSavingGoal(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid id", nil)
		return
	}
	var req dto.ContributeSavingGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	view, err := h.services.SavingGoal.Contribute(c.Request.Context(), userID, id, req.Amount)
	if err != nil {
		response.FromError(c, err)
		return
	}
	resp := dto.SavingGoalResponseFrom(&view.SavingGoal)
	resp.ProgressPercent = view.ProgressPercent
	response.Success(c, http.StatusOK, resp)
}

func (h *Handler) deleteSavingGoal(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid id", nil)
		return
	}
	if err := h.services.SavingGoal.Delete(c.Request.Context(), userID, id); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"deleted": true})
}
