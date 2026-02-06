package ports

import (
	"context"

	"github.com/codevshl/http-metadata-inventory-service/internal/core/domain"
)

// MetadataService defines the primary port for metadata operations.
type MetadataService interface {
	// CreateMetadata initiates the collection of metadata for a URL.
	CreateMetadata(ctx context.Context, url string) error

	// GetMetadata retrieves metadata for a URL.
	// If metadata is present, it returns it.
	// If not present, it initiates collection and returns nil (signaling accepted).
	GetMetadata(ctx context.Context, url string) (*domain.Metadata, error)
}
