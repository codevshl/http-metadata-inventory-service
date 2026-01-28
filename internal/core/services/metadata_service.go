package services

import (
	"context"
	"sync"

	"github.com/codevshl/http-metadata-inventory-service/internal/core/apperror"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/domain"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/ports"
	"github.com/codevshl/http-metadata-inventory-service/internal/platform/logger"
	"go.uber.org/zap"
)

// scrapeTask represents a background scraping task
type scrapeTask struct {
	url string
}

type metadataService struct {
	repo          ports.MetadataRepository
	scraper       ports.Scraper
	taskQueue     chan scrapeTask
	workerPool    *sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	workerCount   int
	queueCapacity int
}

// Config holds configuration for the metadata service
type Config struct {
	WorkerPoolSize int
	TaskQueueSize  int
}

// NewMetadataService creates a new instance of the metadata service with worker pool
func NewMetadataService(repo ports.MetadataRepository, scraper ports.Scraper, cfg Config) ports.MetadataService {
	if cfg.WorkerPoolSize <= 0 {
		cfg.WorkerPoolSize = 10
	}
	if cfg.TaskQueueSize <= 0 {
		cfg.TaskQueueSize = 100
	}

	ctx, cancel := context.WithCancel(context.Background())

	svc := &metadataService{
		repo:          repo,
		scraper:       scraper,
		taskQueue:     make(chan scrapeTask, cfg.TaskQueueSize),
		workerPool:    &sync.WaitGroup{},
		ctx:           ctx,
		cancel:        cancel,
		workerCount:   cfg.WorkerPoolSize,
		queueCapacity: cfg.TaskQueueSize,
	}

	svc.startWorkers()

	logger.Info("Metadata service initialized",
		zap.Int("workers", cfg.WorkerPoolSize),
		zap.Int("queue_capacity", cfg.TaskQueueSize))

	return svc
}

// startWorkers launches the worker pool goroutines
func (s *metadataService) startWorkers() {
	for i := 0; i < s.workerCount; i++ {
		s.workerPool.Add(1)
		go s.worker(i)
	}
}

// worker processes scraping tasks from the queue
func (s *metadataService) worker(id int) {
	defer s.workerPool.Done()

	logger.Debug("Worker started", zap.Int("worker_id", id))

	for {
		select {
		case <-s.ctx.Done():
			logger.Debug("Worker shutting down", zap.Int("worker_id", id))
			return
		case task, ok := <-s.taskQueue:
			if !ok {
				logger.Debug("Task queue closed, worker exiting", zap.Int("worker_id", id))
				return
			}

			s.processTask(task, id)
		}
	}
}

// processTask handles the actual scraping and saving
func (s *metadataService) processTask(task scrapeTask, workerID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*1000000000)
	defer cancel()

	logger.Info("Processing scrape task",
		zap.Int("worker_id", workerID),
		zap.String("url", task.url))

	metadata, err := s.scraper.Fetch(ctx, task.url)
	if err != nil {
		logger.Error("Failed to scrape URL",
			zap.Int("worker_id", workerID),
			zap.String("url", task.url),
			zap.Error(err))
		return
	}

	if err := s.repo.Save(ctx, *metadata); err != nil {
		logger.Error("Failed to save metadata",
			zap.Int("worker_id", workerID),
			zap.String("url", task.url),
			zap.Error(err))
		return
	}

	logger.Info("Successfully processed scrape task",
		zap.Int("worker_id", workerID),
		zap.String("url", task.url))
}

// CreateMetadata synchronously scrapes and saves metadata for a URL
func (s *metadataService) CreateMetadata(ctx context.Context, url string) error {
	if url == "" {
		return domain.ErrInvalidURL
	}

	logger.Info("Creating metadata", zap.String("url", url))

	metadata, err := s.scraper.Fetch(ctx, url)
	if err != nil {
		return apperror.Wrap(apperror.ErrUpstream, "failed to fetch metadata", err)
	}

	if err := s.repo.Save(ctx, *metadata); err != nil {
		return apperror.Wrap(apperror.ErrInternal, "internal server error", err)
	}

	logger.Info("Metadata created successfully", zap.String("url", url))
	return nil
}

// GetMetadata retrieves metadata for a URL, triggering background scraping if not found
func (s *metadataService) GetMetadata(ctx context.Context, url string) (*domain.Metadata, error) {
	if url == "" {
		return nil, domain.ErrInvalidURL
	}

	logger.Info("Getting metadata", zap.String("url", url))

	metadata, err := s.repo.Get(ctx, url)
	if err == nil {
		logger.Info("Metadata found in database", zap.String("url", url))
		return metadata, nil
	}

	if apperror.Is(err, apperror.ErrNotFound) {
		logger.Info("Metadata not found, queueing background scrape", zap.String("url", url))

		select {
		case s.taskQueue <- scrapeTask{url: url}:
			logger.Debug("Task queued successfully", zap.String("url", url))
		default:
			logger.Warn("Task queue is full, scrape may be delayed",
				zap.String("url", url),
				zap.Int("queue_capacity", s.queueCapacity))
		}

		return nil, nil
	}

	return nil, apperror.Wrap(apperror.ErrInternal, "internal server error", err)
}

// Shutdown gracefully stops the worker pool
func (s *metadataService) Shutdown() {
	logger.Info("Shutting down metadata service")

	s.cancel()

	close(s.taskQueue)

	s.workerPool.Wait()

	logger.Info("Metadata service shutdown complete")
}
