package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/diyas/fintrack/internal/delivery/dto"
	"github.com/diyas/fintrack/internal/delivery/middleware"
	"github.com/diyas/fintrack/pkg/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initCSVRoutes(router *gin.RouterGroup) {
	auth := middleware.Auth(h.services.Auth.TokenManager())
	router.GET("/export/transactions", auth, h.exportTransactions)
	router.POST("/import/transactions", auth, h.importTransactions)
}

func (h *Handler) exportTransactions(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	filename := fmt.Sprintf("transactions-%s.csv", time.Now().UTC().Format("20060102"))
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if err := h.services.CSV.Export(c.Request.Context(), userID, c.Writer); err != nil {
		response.FromError(c, err)
		return
	}
}

func (h *Handler) importTransactions(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	header, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "file form field required", nil)
		return
	}
	file, err := header.Open()
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "cannot open uploaded file", nil)
		return
	}
	defer file.Close()
	summary, err := h.services.CSV.Import(c.Request.Context(), userID, file)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.CSVImportResponseFrom(summary))
}
