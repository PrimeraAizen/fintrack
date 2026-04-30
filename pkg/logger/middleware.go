package logger

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RequestIDKey     = "request_id"
	CorrelationIDKey = "correlation_id"
	UserIDKey        = "user_id"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// LoggingMiddleware logs HTTP requests and responses
func LoggingMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Create context with request ID
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, c.GetString(RequestIDKey))
		c.Request = c.Request.WithContext(ctx)

		// Log request
		logger.WithContext(ctx).
			WithRequest(c.Request.Method, c.Request.URL.Path).
			WithFields(Fields{
				"user_agent":     c.Request.UserAgent(),
				"remote_addr":    c.ClientIP(),
				"content_type":   c.Request.Header.Get("Content-Type"),
				"content_length": c.Request.ContentLength,
			}).
			Info("HTTP request started")

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log response
		logger.WithContext(ctx).
			WithRequest(c.Request.Method, c.Request.URL.Path).
			WithResponse(c.Writer.Status(), int64(c.Writer.Size())).
			WithDuration(duration).
			WithFields(Fields{
				"response_time_ms": duration.Milliseconds(),
			}).
			Info("HTTP request completed")
	}
}

// RecoveryMiddleware recovers from panics and logs them
func RecoveryMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				logger.WithContext(c.Request.Context()).
					WithRequest(c.Request.Method, c.Request.URL.Path).
					WithFields(Fields{
						"panic": err,
						"stack": getStackTrace(),
					}).
					Error("Panic recovered")

				// Return 500 error
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":      "Internal server error",
					"request_id": c.GetString(RequestIDKey),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Simple request ID generation - in production you might want to use UUID
	return strconv.FormatInt(time.Now().UnixNano(), 36) + strconv.FormatInt(rand.Int63(), 36)
}

// getStackTrace returns a stack trace (simplified version)
func getStackTrace() string {
	// This is a simplified version - in production you might want to use
	// a more sophisticated stack trace library
	return "stack trace not implemented"
}

// ContextMiddleware adds logger to Gin context
func ContextMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), "logger", logger)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// GetLoggerFromContext retrieves logger from context
func GetLoggerFromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value("logger").(*Logger); ok {
		return logger
	}
	return Default()
}

// SetUserID sets user ID in context for logging
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// SetCorrelationID sets correlation ID in context for logging
func SetCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}
