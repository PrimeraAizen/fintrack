package v1

import (
	"net/http"
	"time"

	"github.com/diyas/fintrack/internal/delivery/dto"
	"github.com/diyas/fintrack/internal/delivery/middleware"
	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) initTransactionRoutes(router *gin.RouterGroup) {
	tx := router.Group("/transactions", middleware.Auth(h.services.Auth.TokenManager()))
	{
		tx.POST("", h.createTransaction)
		tx.GET("", h.listTransactions)
		tx.GET("/:id", h.getTransaction)
		tx.PATCH("/:id", h.updateTransaction)
		tx.DELETE("/:id", h.deleteTransaction)
	}
}

func (h *Handler) createTransaction(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	result, err := h.services.Transaction.Create(c.Request.Context(), userID, req.ToInput())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, gin.H{
		"transaction":     dto.TransactionResponseFrom(result.Transaction),
		"budget_exceeded": result.BudgetExceeded,
	})
}

func (h *Handler) listTransactions(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	filter, err := parseTransactionFilter(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	transactions, total, err := h.services.Transaction.List(c.Request.Context(), userID, filter)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.PaginatedSuccess(c, dto.TransactionListResponseFrom(transactions), filter.Page, filter.PerPage, total)
}

func (h *Handler) getTransaction(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid id", nil)
		return
	}
	t, err := h.services.Transaction.Get(c.Request.Context(), userID, id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.TransactionResponseFrom(t))
}

func (h *Handler) updateTransaction(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid id", nil)
		return
	}
	var req dto.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	t, err := h.services.Transaction.Update(c.Request.Context(), userID, id, req.ToInput())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.TransactionResponseFrom(t))
}

func (h *Handler) deleteTransaction(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid id", nil)
		return
	}
	if err := h.services.Transaction.Delete(c.Request.Context(), userID, id); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"deleted": true})
}

func parseTransactionFilter(c *gin.Context) (domain.TransactionFilter, error) {
	pageParams := domain.PageParams{
		Page:    parseIntQuery(c, "page"),
		PerPage: parseIntQuery(c, "per_page"),
	}.Normalize()

	filter := domain.TransactionFilter{
		Page:    pageParams.Page,
		PerPage: pageParams.PerPage,
		Type:    c.Query("type"),
	}
	if v := c.Query("account_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return filter, err
		}
		filter.AccountID = &id
	}
	if v := c.Query("category_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return filter, err
		}
		filter.CategoryID = &id
	}
	if v := c.Query("from_date"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return filter, err
		}
		filter.FromDate = &t
	}
	if v := c.Query("to_date"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return filter, err
		}
		filter.ToDate = &t
	}
	return filter, nil
}

func parseIntQuery(c *gin.Context, key string) int {
	v := c.Query(key)
	if v == "" {
		return 0
	}
	n := 0
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return 0
		}
		n = n*10 + int(ch-'0')
	}
	return n
}
