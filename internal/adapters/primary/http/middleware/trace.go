package middleware

import (
	"github.com/codevshl/http-metadata-inventory-service/internal/adapters/primary/http/response"
	"github.com/codevshl/http-metadata-inventory-service/internal/platform/logger"
	"github.com/gin-gonic/gin"
)

// TraceMiddleware injects a trace ID into the request context and response headers.
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := response.ReadTraceID(c)
		if traceID == "" {
			traceID = response.NewTraceID()
		}
	if traceID != "" {
		response.SetTraceID(c, traceID)
	}

	ctx := logger.ContextWithReqID(c.Request.Context(), traceID)
	c.Request = c.Request.WithContext(ctx)

	c.Next()
}
}
