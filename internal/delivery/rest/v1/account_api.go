package v1

import (
	"net/http"

	"github.com/diyas/fintrack/internal/delivery/dto"
	"github.com/diyas/fintrack/internal/delivery/middleware"
	"github.com/diyas/fintrack/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) initAccountRoutes(router *gin.RouterGroup) {
	accounts := router.Group("/accounts", middleware.Auth(h.services.Auth.TokenManager()))
	{
		accounts.POST("", h.createAccount)
		accounts.GET("", h.listAccounts)
		accounts.GET("/:id", h.getAccount)
		accounts.PUT("/:id", h.updateAccount)
		accounts.DELETE("/:id", h.deleteAccount)
	}
}

func (h *Handler) createAccount(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	var req dto.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	account, err := h.services.Account.Create(c.Request.Context(), userID, req.ToInput())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, dto.AccountResponseFrom(account))
}

func (h *Handler) listAccounts(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	accounts, err := h.services.Account.List(c.Request.Context(), userID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.AccountListResponseFrom(accounts))
}

func (h *Handler) getAccount(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid id", nil)
		return
	}
	account, err := h.services.Account.Get(c.Request.Context(), userID, id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.AccountResponseFrom(account))
}

func (h *Handler) updateAccount(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid id", nil)
		return
	}
	var req dto.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	account, err := h.services.Account.Update(c.Request.Context(), userID, id, req.ToInput())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.AccountResponseFrom(account))
}

func (h *Handler) deleteAccount(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid id", nil)
		return
	}
	if err := h.services.Account.Delete(c.Request.Context(), userID, id); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"deleted": true})
}
