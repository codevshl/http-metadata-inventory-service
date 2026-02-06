package v1

import (
	"net/http"

	"github.com/codevshl/http-metadata-inventory-service/internal/adapters/primary/http/response"
	"github.com/codevshl/http-metadata-inventory-service/internal/platform/logger"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles GET /health
// @Summary Health check
// @Description Returns a simple status response indicating the service is healthy.
// @Tags Health
// @Success 200 {object} response.APIResponse "success=true"
// @Router /health [get]
func HealthHandler(c *gin.Context) {
	logger.WithRequest(c.Request.Context()).Debug(
		logger.Msg("HTTP", "HealthHandler", "Adapter", "health check"),
	)
	response.WriteSuccess(c, http.StatusOK, "HEALTH_OK", "ok", nil)
}
