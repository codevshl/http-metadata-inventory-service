package http

import (
	"net/http"

	"github.com/codevshl/http-metadata-inventory-service/internal/core/domain"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/ports"
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.svc.CreateMetadata(c.Request.Context(), req.URL); err != nil {
		switch err {
		case domain.ErrInvalidURL:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.Status(http.StatusCreated)
}

// GetMetadataHandler handles GET /metadata
func (h *Handler) GetMetadataHandler(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "url query parameter is required"})
		return
	}

	metadata, err := h.svc.GetMetadata(c.Request.Context(), url)
	if err != nil {
		switch err {
		case domain.ErrNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	if metadata == nil {
		c.Status(http.StatusAccepted)
		return
	}

	c.JSON(http.StatusOK, MetadataResponse{
		URL:        metadata.URL,
		Headers:    metadata.Headers,
		Cookies:    metadata.Cookies,
		PageSource: metadata.PageSource,
	})
}
