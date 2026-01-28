package middleware

import (
	"net/http"
	"time"

	"github.com/codevshl/http-metadata-inventory-service/internal/adapters/primary/http/response"
	"github.com/codevshl/http-metadata-inventory-service/internal/platform/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// JSONMiddleware ensures that all responses are JSON
func JSONMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Next()
	}
}

// ErrorHandlerMiddleware handles errors and formats them consistently
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// If there are errors in the context, handle the last one
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			logger.Error("Request Error", zap.Error(err.Err))

			// If not already written, write error response
			if !c.Writer.Written() {
				response.WriteError(c, err.Err)
			}
		}
	}
}

// RequestLogger logs incoming requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
		)
	}
}

// RecoveryMiddleware recovers from panics and logs them
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(error); ok {
			logger.Error("Panic Recovered", zap.Error(err))
		} else {
			logger.Error("Panic Recovered", zap.Any("error", recovered))
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	})
}
