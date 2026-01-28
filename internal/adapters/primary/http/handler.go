package http

import (
	"net/http"

	"github.com/codevshl/http-metadata-inventory-service/internal/core/apperror"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/ports"
	"github.com/codevshl/http-metadata-inventory-service/internal/adapters/primary/http/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc ports.MetadataService
}

func NewHandler(svc ports.MetadataService) *Handler {
	return &Handler{
		svc: svc,
	}
}

// CreateMetadataHandler handles POST /metadata
func (h *Handler) CreateMetadataHandler(c *gin.Context) {
	var req CreateMetadataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteError(c, apperror.Wrap(apperror.ErrValidation, "invalid request body", err))
		return
	}

	if err := h.svc.CreateMetadata(c.Request.Context(), req.URL); err != nil {
		response.WriteError(c, err)
		return
	}

	response.WriteSuccess(c, http.StatusCreated, "METADATA_CREATED", "metadata created", nil)
}

// GetMetadataHandler handles GET /metadata
func (h *Handler) GetMetadataHandler(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		response.WriteError(c, apperror.New(apperror.ErrValidation, "url query parameter is required"))
		return
	}

	metadata, err := h.svc.GetMetadata(c.Request.Context(), url)
	if err != nil {
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
