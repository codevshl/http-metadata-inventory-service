package v1

import (
	"net/http"

	"github.com/codevshl/http-metadata-inventory-service/internal/adapters/primary/http/response"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/apperror"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/ports"
	"github.com/codevshl/http-metadata-inventory-service/internal/platform/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MetadataHandler struct {
	svc ports.MetadataService
}

func NewMetadataHandler(svc ports.MetadataService) *MetadataHandler {
	return &MetadataHandler{svc: svc}
}

// CreateMetadataHandler handles POST /metadata
// @Summary Create metadata (synchronous scrape)
// @Description Scrapes the URL immediately and stores the metadata.
// @Tags Metadata
// @Accept json
// @Produce json
// @Param request body CreateMetadataRequest true "Create metadata request"
// @Success 201 {object} response.APIResponse "success=true"
// @Failure 400 {object} response.APIResponse "success=false"
// @Failure 500 {object} response.APIResponse "success=false"
// @Failure 502 {object} response.APIResponse "success=false"
// @Router /metadata [post]
func (h *MetadataHandler) CreateMetadataHandler(c *gin.Context) {
	var req CreateMetadataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithRequest(c.Request.Context(), zap.Error(err)).Warn(
			logger.Msg("HTTP", "MetadataHandler", "Adapter", "invalid request body"),
		)
		response.WriteError(c, apperror.Wrap(apperror.ErrValidation, "invalid request body", err))
		return
	}

	logger.WithRequest(c.Request.Context(), zap.String("url", req.URL)).Info(
		logger.Msg("HTTP", "MetadataHandler", "Adapter", "create metadata requested"),
	)

	if err := h.svc.CreateMetadata(c.Request.Context(), req.URL); err != nil {
		logger.WithRequest(c.Request.Context(), zap.Error(err), zap.String("url", req.URL)).Error(
			logger.Msg("HTTP", "MetadataHandler", "Adapter", "create metadata failed"),
		)
		response.WriteError(c, err)
		return
	}

	response.WriteSuccess(c, http.StatusCreated, "METADATA_CREATED", "metadata created", nil)
}

// GetMetadataHandler handles GET /metadata
// @Summary Get metadata (cached or async)
// @Description Returns metadata if available, otherwise starts a background scrape and returns 202.
// @Tags Metadata
// @Accept json
// @Produce json
// @Param url query string true "Target URL" format(uri)
// @Success 200 {object} response.APIResponse{data=MetadataResponse} "success=true"
// @Success 202 {object} response.APIResponse "success=true"
// @Failure 400 {object} response.APIResponse "success=false"
// @Failure 500 {object} response.APIResponse "success=false"
// @Router /metadata [get]
func (h *MetadataHandler) GetMetadataHandler(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		logger.WithRequest(c.Request.Context()).Warn(
			logger.Msg("HTTP", "MetadataHandler", "Adapter", "missing url query parameter"),
		)
		response.WriteError(c, apperror.New(apperror.ErrValidation, "url query parameter is required"))
		return
	}

	logger.WithRequest(c.Request.Context(), zap.String("url", url)).Info(
		logger.Msg("HTTP", "MetadataHandler", "Adapter", "get metadata requested"),
	)

	metadata, err := h.svc.GetMetadata(c.Request.Context(), url)
	if err != nil {
		logger.WithRequest(c.Request.Context(), zap.Error(err), zap.String("url", url)).Error(
			logger.Msg("HTTP", "MetadataHandler", "Adapter", "get metadata failed"),
		)
		response.WriteError(c, err)
		return
	}

	if metadata == nil {
		response.WriteSuccess(c, http.StatusAccepted, "METADATA_PENDING", "metadata collection started", nil)
		return
	}

	response.WriteSuccess(c, http.StatusOK, "METADATA_FOUND", "metadata retrieved", MetadataResponse{
		URL:        metadata.URL,
		Headers:    metadata.Headers,
		Cookies:    metadata.Cookies,
		PageSource: metadata.PageSource,
	})
}
