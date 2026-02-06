package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/codevshl/http-metadata-inventory-service/internal/core/apperror"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/domain"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, metadata domain.Metadata) error {
	args := m.Called(ctx, metadata)
	return args.Error(0)
}

func (m *MockRepository) Get(ctx context.Context, url string) (*domain.Metadata, error) {
	args := m.Called(ctx, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Metadata), args.Error(1)
}

type MockScraper struct {
	mock.Mock
}

func (m *MockScraper) Fetch(ctx context.Context, url string) (*domain.Metadata, error) {
	args := m.Called(ctx, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Metadata), args.Error(1)
}

func TestCreateMetadata_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockScraper := new(MockScraper)

	testURL := "https://example.com"
	testMetadata := &domain.Metadata{
		URL:     testURL,
		Headers: map[string]string{"Content-Type": "text/html"},
		Cookies: map[string]string{},
	}

	mockScraper.On("Fetch", mock.Anything, testURL).Return(testMetadata, nil)
	mockRepo.On("Save", mock.Anything, *testMetadata).Return(nil)

	svc := services.NewMetadataService(mockRepo, mockScraper, services.Config{
		WorkerPoolSize: 1,
		TaskQueueSize:  10,
	})
	defer svc.(interface{ Shutdown() }).Shutdown()

	err := svc.CreateMetadata(context.Background(), testURL)

	assert.NoError(t, err)
	mockScraper.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestCreateMetadata_InvalidURL(t *testing.T) {
	mockRepo := new(MockRepository)
	mockScraper := new(MockScraper)

	svc := services.NewMetadataService(mockRepo, mockScraper, services.Config{
		WorkerPoolSize: 1,
		TaskQueueSize:  10,
	})
	defer svc.(interface{ Shutdown() }).Shutdown()

	err := svc.CreateMetadata(context.Background(), "")

	assert.True(t, apperror.Is(err, apperror.ErrValidation))
}

func TestCreateMetadata_ScraperError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockScraper := new(MockScraper)

	testURL := "https://example.com"
	scraperErr := errors.New("network error")

	mockScraper.On("Fetch", mock.Anything, testURL).Return(nil, scraperErr)

	svc := services.NewMetadataService(mockRepo, mockScraper, services.Config{
		WorkerPoolSize: 1,
		TaskQueueSize:  10,
	})
	defer svc.(interface{ Shutdown() }).Shutdown()

	err := svc.CreateMetadata(context.Background(), testURL)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch metadata")
	mockScraper.AssertExpectations(t)
}

func TestGetMetadata_Found(t *testing.T) {
	mockRepo := new(MockRepository)
	mockScraper := new(MockScraper)

	testURL := "https://example.com"
	testMetadata := &domain.Metadata{
		URL:     testURL,
		Headers: map[string]string{"Content-Type": "text/html"},
	}

	mockRepo.On("Get", mock.Anything, testURL).Return(testMetadata, nil)

	svc := services.NewMetadataService(mockRepo, mockScraper, services.Config{
		WorkerPoolSize: 1,
		TaskQueueSize:  10,
	})
	defer svc.(interface{ Shutdown() }).Shutdown()

	result, err := svc.GetMetadata(context.Background(), testURL)

	assert.NoError(t, err)
	assert.Equal(t, testMetadata, result)
	mockRepo.AssertExpectations(t)
}

func TestGetMetadata_NotFound_QueuesTask(t *testing.T) {
	mockRepo := new(MockRepository)
	mockScraper := new(MockScraper)

	testURL := "https://example.com"
	testMetadata := &domain.Metadata{
		URL:     testURL,
		Headers: map[string]string{"Content-Type": "text/html"},
	}

	mockRepo.On("Get", mock.Anything, testURL).Return(nil, domain.ErrNotFound)
	mockScraper.On("Fetch", mock.Anything, testURL).Return(testMetadata, nil).Maybe()
	mockRepo.On("Save", mock.Anything, *testMetadata).Return(nil).Maybe()

	svc := services.NewMetadataService(mockRepo, mockScraper, services.Config{
		WorkerPoolSize: 1,
		TaskQueueSize:  10,
	})
	defer svc.(interface{ Shutdown() }).Shutdown()

	result, err := svc.GetMetadata(context.Background(), testURL)

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertCalled(t, "Get", mock.Anything, testURL)
}

func TestGetMetadata_InvalidURL(t *testing.T) {
	mockRepo := new(MockRepository)
	mockScraper := new(MockScraper)

	svc := services.NewMetadataService(mockRepo, mockScraper, services.Config{
		WorkerPoolSize: 1,
		TaskQueueSize:  10,
	})
	defer svc.(interface{ Shutdown() }).Shutdown()

	result, err := svc.GetMetadata(context.Background(), "")

	assert.True(t, apperror.Is(err, apperror.ErrValidation))
	assert.Nil(t, result)
}
