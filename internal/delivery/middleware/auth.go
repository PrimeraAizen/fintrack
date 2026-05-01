package middleware

import (
	"net/http"
	"strings"

	"github.com/diyas/fintrack/internal/service"
	"github.com/diyas/fintrack/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const UserIDKey = "userID"

func Auth(tm service.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "missing authorization header", nil)
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "invalid authorization header", nil)
			return
		}

		claims, err := tm.ParseAccess(parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "invalid or expired token", nil)
			return
		}

		if claims.ID != "" {
			if blacklisted, _ := tm.IsAccessBlacklisted(c.Request.Context(), claims.ID); blacklisted {
				response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "token has been revoked", nil)
				return
			}
		}

		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}

func UserIDFrom(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(UserIDKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}
