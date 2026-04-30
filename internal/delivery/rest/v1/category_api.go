package v1

import (
	"net/http"

	"github.com/diyas/fintrack/internal/delivery/dto"
	"github.com/diyas/fintrack/internal/delivery/middleware"
	"github.com/diyas/fintrack/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) initCategoryRoutes(router *gin.RouterGroup) {
	categories := router.Group("/categories", middleware.Auth(h.services.Auth.TokenManager()))
	{
		categories.POST("", h.createCategory)
		categories.GET("", h.listCategories)
		categories.PUT("/:id", h.updateCategory)
		categories.DELETE("/:id", h.deleteCategory)
	}
}

func (h *Handler) createCategory(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	category, err := h.services.Category.Create(c.Request.Context(), userID, req.ToInput())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, dto.CategoryResponseFrom(category))
}

func (h *Handler) listCategories(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	typeFilter := c.Query("type")
	categories, err := h.services.Category.List(c.Request.Context(), userID, typeFilter)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.CategoryListResponseFrom(categories))
}

func (h *Handler) updateCategory(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid id", nil)
		return
	}
	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	category, err := h.services.Category.Update(c.Request.Context(), userID, id, req.ToInput())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.CategoryResponseFrom(category))
}

func (h *Handler) deleteCategory(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid id", nil)
		return
	}
	if err := h.services.Category.Delete(c.Request.Context(), userID, id); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"deleted": true})
}
