package ports

import (
	"context"

	"github.com/codevshl/http-metadata-inventory-service/internal/core/domain"
)

// MetadataRepository defines the secondary port for data persistence.
type MetadataRepository interface {
	Save(ctx context.Context, metadata domain.Metadata) error
	Get(ctx context.Context, url string) (*domain.Metadata, error)
}

// Scraper defines the secondary port for external content fetching.
type Scraper interface {
	Fetch(ctx context.Context, url string) (*domain.Metadata, error)
}
