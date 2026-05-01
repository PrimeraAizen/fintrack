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

func (h *Handler) initReportRoutes(router *gin.RouterGroup) {
	reports := router.Group("/reports", middleware.Auth(h.services.Auth.TokenManager()))
	{
		reports.GET("/weekly", h.weeklyReport)
		reports.GET("/monthly", h.monthlyReport)
	}
}

func (h *Handler) weeklyReport(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	anchor := time.Now().UTC()
	if d := c.Query("weekOf"); d != "" {
		parsed, err := time.Parse("2006-01-02", d)
		if err != nil {
			response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid weekOf, expected YYYY-MM-DD", nil)
			return
		}
		anchor = parsed
	}
	r, err := h.services.Report.Weekly(c.Request.Context(), userID, anchor)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ReportResponseFrom(r))
}

func (h *Handler) monthlyReport(c *gin.Context) {
	userID, _ := middleware.UserIDFrom(c)
	now := time.Now().UTC()
	year := now.Year()
	month := int(now.Month())
	if m := c.Query("month"); m != "" {
		// Accept YYYY-MM format.
		t, err := time.Parse("2006-01", m)
		if err != nil {
			response.Error(c, http.StatusBadRequest, response.CodeBadRequest,
				fmt.Sprintf("invalid month %q, expected YYYY-MM", m), nil)
			return
		}
		year = t.Year()
		month = int(t.Month())
	}
	r, err := h.services.Report.Monthly(c.Request.Context(), userID, year, time.Month(month))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ReportResponseFrom(r))
}
