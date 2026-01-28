package mongodb

import (
	"time"

	"github.com/codevshl/http-metadata-inventory-service/internal/core/domain"
)

// mongoMetadata is the internal representation for MongoDB persistence
type mongoMetadata struct {
	URL        string            `bson:"url"`
	Headers    map[string]string `bson:"headers"`
	Cookies    map[string]string `bson:"cookies"`
	PageSource string            `bson:"page_source"`
	CreatedAt  time.Time         `bson:"created_at"`
	UpdatedAt  time.Time         `bson:"updated_at"`
}

// toMongoModel converts domain entity to MongoDB model
func toMongoModel(m domain.Metadata) mongoMetadata {
	now := time.Now()
	return mongoMetadata{
		URL:        m.URL,
		Headers:    m.Headers,
		Cookies:    m.Cookies,
		PageSource: m.PageSource,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// toDomain converts MongoDB model to domain entity
func toDomain(m mongoMetadata) domain.Metadata {
	return domain.Metadata{
		URL:        m.URL,
		Headers:    m.Headers,
		Cookies:    m.Cookies,
		PageSource: m.PageSource,
	}
}
