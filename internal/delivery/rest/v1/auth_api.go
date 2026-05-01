package v1

import (
	"net/http"
	"strings"

	"github.com/diyas/fintrack/internal/delivery/dto"
	"github.com/diyas/fintrack/internal/delivery/middleware"
	"github.com/diyas/fintrack/pkg/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initAuthRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.register)
		auth.POST("/login", h.login)
		auth.POST("/refresh", h.refresh)
		auth.POST("/logout", middleware.Auth(h.services.Auth.TokenManager()), h.logout)
		auth.GET("/me", middleware.Auth(h.services.Auth.TokenManager()), h.me)
	}
	// Canonical path expected by the iOS bootstrap call.
	router.GET("/me", middleware.Auth(h.services.Auth.TokenManager()), h.me)
}

func (h *Handler) register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	tokens, user, err := h.services.Auth.Register(c.Request.Context(), req.ToInput())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, gin.H{
		"user":   dto.UserResponseFrom(user),
		"tokens": dto.TokenResponseFrom(tokens),
	})
}

func (h *Handler) login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	tokens, err := h.services.Auth.Login(c.Request.Context(), req.ToInput())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.TokenResponseFrom(tokens))
}

func (h *Handler) refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	tokens, err := h.services.Auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.TokenResponseFrom(tokens))
}

func (h *Handler) logout(c *gin.Context) {
	rawToken := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	var req dto.LogoutRequest
	_ = c.ShouldBindJSON(&req) // body is optional
	if err := h.services.Auth.Logout(c.Request.Context(), rawToken, req.RefreshToken); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"logged_out": true})
}

func (h *Handler) me(c *gin.Context) {
	userID, ok := middleware.UserIDFrom(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "missing user", nil)
		return
	}
	user, err := h.services.Auth.Me(c.Request.Context(), userID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.UserResponseFrom(user))
}
