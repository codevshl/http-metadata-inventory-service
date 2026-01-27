package http

import (
	"github.com/codevshl/http-metadata-inventory-service/internal/adapters/primary/http/middleware"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/ports"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures the Gin engine and routes
func SetupRouter(svc ports.MetadataService, isProd bool) *gin.Engine {
	if isProd {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Apply Middlewares
	r.Use(middleware.RequestLogger())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.JSONMiddleware())
	r.Use(middleware.ErrorHandlerMiddleware())

	// Handlers
	h := NewHandler(svc)

	// Routes
	v1 := r.Group("/api/v1")
	{
		v1.POST("/metadata", h.CreateMetadataHandler)
		v1.GET("/metadata", h.GetMetadataHandler)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
