package response

import (
	"errors"
	"net/http"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/gin-gonic/gin"
)

// Error code constants sent in the "code" field of every error response.
const (
	CodeBadRequest       = "bad_request"
	CodeUnauthorized     = "unauthorized"
	CodeForbidden        = "forbidden"
	CodeNotFound         = "not_found"
	CodeConflict         = "conflict"
	CodeValidationFailed = "validation_failed"
	CodeInternalError    = "internal_error"
	CodeRateLimited      = "rate_limited"
)

type ErrorBody struct {
	Error   bool              `json:"error"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type SuccessBody struct {
	Error bool        `json:"error"`
	Data  interface{} `json:"data"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedData struct {
	Items      interface{} `json:"items"`
	Pagination Pagination  `json:"pagination"`
}

func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, SuccessBody{Error: false, Data: data})
}

func Error(c *gin.Context, status int, code, message string, details map[string]string) {
	c.AbortWithStatusJSON(status, ErrorBody{Error: true, Code: code, Message: message, Details: details})
}

func ValidationError(c *gin.Context, details map[string]string) {
	Error(c, http.StatusBadRequest, CodeValidationFailed, "Validation failed", details)
}

func FromError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		Error(c, http.StatusNotFound, CodeNotFound, err.Error(), nil)
	case errors.Is(err, domain.ErrUnauthorized),
		errors.Is(err, domain.ErrInvalidCredentials),
		errors.Is(err, domain.ErrInvalidToken):
		Error(c, http.StatusUnauthorized, CodeUnauthorized, err.Error(), nil)
	case errors.Is(err, domain.ErrForbidden):
		Error(c, http.StatusForbidden, CodeForbidden, err.Error(), nil)
	case errors.Is(err, domain.ErrConflict):
		Error(c, http.StatusConflict, CodeConflict, err.Error(), nil)
	case errors.Is(err, domain.ErrInsufficientBalance),
		errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrValidation):
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error(), nil)
	default:
		Error(c, http.StatusInternalServerError, CodeInternalError, err.Error(), nil)
	}
}

func PaginatedSuccess(c *gin.Context, items interface{}, page, perPage int, total int64) {
	totalPages := 0
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	Success(c, http.StatusOK, PaginatedData{
		Items: items,
		Pagination: Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
