package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/codevshl/http-metadata-inventory-service/internal/adapters/secondary/repository/mongodb"
	"github.com/codevshl/http-metadata-inventory-service/internal/adapters/secondary/scraper/http_client"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/apperror"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/domain"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/services"
	"github.com/codevshl/http-metadata-inventory-service/internal/platform/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testMongoURI = "mongodb://localhost:27017"
	testDBName   = "metadata_inventory_test"
)

func setupTestDB(t *testing.T) {
	err := db.Connect(testMongoURI)
	require.NoError(t, err, "Failed to connect to test MongoDB")
}

func teardownTestDB(t *testing.T) {
	client := db.GetClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Database(testDBName).Drop(ctx)
	if err != nil {
		t.Logf("Warning: Failed to drop test database: %v", err)
	}

	db.Disconnect()
}

func TestIntegration_CreateAndGetMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	repo := mongodb.NewMongoRepository(db.GetClient(), testDBName)
	scraper := http_client.NewScraper()

	svc := services.NewMetadataService(repo, scraper, services.Config{
		WorkerPoolSize: 2,
		TaskQueueSize:  10,
	})
	defer svc.(interface{ Shutdown() }).Shutdown()

	ctx := context.Background()
	testURL := "https://example.com"

	err := svc.CreateMetadata(ctx, testURL)
	require.NoError(t, err, "CreateMetadata should succeed")

	metadata, err := svc.GetMetadata(ctx, testURL)
	require.NoError(t, err, "GetMetadata should succeed")
	require.NotNil(t, metadata, "Metadata should not be nil")

	assert.Equal(t, testURL, metadata.URL)
	assert.NotEmpty(t, metadata.Headers, "Headers should not be empty")
	assert.NotEmpty(t, metadata.PageSource, "Page source should not be empty")
}

func TestIntegration_GetMetadata_NotFound_BackgroundScrape(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	repo := mongodb.NewMongoRepository(db.GetClient(), testDBName)
	scraper := http_client.NewScraper()

	svc := services.NewMetadataService(repo, scraper, services.Config{
		WorkerPoolSize: 2,
		TaskQueueSize:  10,
	})
	defer svc.(interface{ Shutdown() }).Shutdown()

	ctx := context.Background()
	testURL := "https://httpbin.org/html"

	metadata, err := svc.GetMetadata(ctx, testURL)
	require.NoError(t, err, "First GetMetadata should succeed")
	assert.Nil(t, metadata, "First call should return nil (202 Accepted)")

	time.Sleep(3 * time.Second)

	metadata, err = svc.GetMetadata(ctx, testURL)
	require.NoError(t, err, "Second GetMetadata should succeed")
	assert.NotNil(t, metadata, "Second call should return metadata")
	assert.Equal(t, testURL, metadata.URL)
}

func TestIntegration_Repository_SaveAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	repo := mongodb.NewMongoRepository(db.GetClient(), testDBName)
	ctx := context.Background()

	testMetadata := domain.Metadata{
		URL:        "https://test.com",
		Headers:    map[string]string{"Content-Type": "text/html"},
		Cookies:    map[string]string{"session": "abc123"},
		PageSource: "<html>test</html>",
	}

	err := repo.Save(ctx, testMetadata)
	require.NoError(t, err, "Save should succeed")

	retrieved, err := repo.Get(ctx, testMetadata.URL)
	require.NoError(t, err, "Get should succeed")
	require.NotNil(t, retrieved, "Retrieved metadata should not be nil")

	assert.Equal(t, testMetadata.URL, retrieved.URL)
	assert.Equal(t, testMetadata.Headers, retrieved.Headers)
	assert.Equal(t, testMetadata.Cookies, retrieved.Cookies)
	assert.Equal(t, testMetadata.PageSource, retrieved.PageSource)
}

func TestIntegration_Repository_Get_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	repo := mongodb.NewMongoRepository(db.GetClient(), testDBName)
	ctx := context.Background()

	metadata, err := repo.Get(ctx, "https://nonexistent.com")
	assert.True(t, apperror.Is(err, apperror.ErrNotFound))
	assert.Nil(t, metadata)
}

func TestIntegration_Scraper_Fetch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	scraper := http_client.NewScraper()
	ctx := context.Background()

	metadata, err := scraper.Fetch(ctx, "https://example.com")
	require.NoError(t, err, "Fetch should succeed")
	require.NotNil(t, metadata, "Metadata should not be nil")

	assert.Equal(t, "https://example.com", metadata.URL)
	assert.NotEmpty(t, metadata.Headers, "Headers should not be empty")
	assert.NotEmpty(t, metadata.PageSource, "Page source should not be empty")
	assert.Contains(t, metadata.PageSource, "Example Domain")
}
