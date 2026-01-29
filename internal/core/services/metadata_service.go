package services

import (
	"context"
	"sync"
	"time"

	"github.com/codevshl/http-metadata-inventory-service/internal/core/apperror"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/domain"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/ports"
	"github.com/codevshl/http-metadata-inventory-service/internal/platform/logger"
	"go.uber.org/zap"
)

// scrapeTask represents a background scraping task
type scrapeTask struct {
	url     string
	traceID string
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

	logger.WithRequest(context.Background(),
		zap.Int("workers", cfg.WorkerPoolSize),
		zap.Int("queue_capacity", cfg.TaskQueueSize),
	).Info(logger.Msg("Metadata", "Service", "Core", "service initialized"))

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
	defer func() {
		if r := recover(); r != nil {
			logger.WithRequest(context.Background(), zap.Int("worker_id", id), zap.Any("panic", r), zap.Stack("stack")).Error(
				logger.Alert(logger.SeverityP0Critical, "Metadata", "Worker", "Core", "panic recovered in worker"))
		}
		s.workerPool.Done()
	}()

	logger.WithRequest(context.Background(), zap.Int("worker_id", id)).Debug(
		logger.Msg("Metadata", "Worker", "Core", "worker started"),
	)

	for {
		select {
		case <-s.ctx.Done():
			logger.WithRequest(context.Background(), zap.Int("worker_id", id)).Debug(
				logger.Msg("Metadata", "Worker", "Core", "worker shutting down"),
			)
			return
		case task, ok := <-s.taskQueue:
			if !ok {
				logger.WithRequest(context.Background(), zap.Int("worker_id", id)).Debug(
					logger.Msg("Metadata", "Worker", "Core", "task queue closed, worker exiting"),
				)
				return
			}

			s.processTask(task, id)
		}
	}
}

// processTask handles the actual scraping and saving
func (s *metadataService) processTask(task scrapeTask, workerID int) {
	baseCtx := context.Background()
	if task.traceID != "" {
		baseCtx = logger.ContextWithReqID(baseCtx, task.traceID)
	}
	ctx, cancel := context.WithTimeout(baseCtx, 60*time.Second)
	defer cancel()

	logger.WithRequest(ctx, zap.Int("worker_id", workerID), zap.String("url", task.url)).Info(
		logger.Msg("Metadata", "Worker", "Core", "processing scrape task"),
	)

	metadata, err := s.scraper.Fetch(ctx, task.url)
	if err != nil {
		logger.WithRequest(ctx, zap.Int("worker_id", workerID), zap.String("url", task.url), zap.Error(err)).Error(
			logger.Alert(logger.SeverityP1High, "Metadata", "Worker", "Core", "failed to scrape url"),
		)
		return
	}

	if err := s.repo.Save(ctx, *metadata); err != nil {
		logger.WithRequest(ctx, zap.Int("worker_id", workerID), zap.String("url", task.url), zap.Error(err)).Error(
			logger.Alert(logger.SeverityP1High, "Metadata", "Worker", "Core", "failed to save metadata"),
		)
		return
	}

	logger.WithRequest(ctx, zap.Int("worker_id", workerID), zap.String("url", task.url)).Info(
		logger.Msg("Metadata", "Worker", "Core", "successfully processed scrape task"),
	)
}

// CreateMetadata synchronously scrapes and saves metadata for a URL
func (s *metadataService) CreateMetadata(ctx context.Context, url string) error {
	if url == "" {
		return apperror.New(apperror.ErrValidation, "url is required")
	}

	logger.WithRequest(ctx, zap.String("url", url)).Info(
		logger.Msg("Metadata", "Service", "Core", "creating metadata"),
	)

	metadata, err := s.scraper.Fetch(ctx, url)
	if err != nil {
		return apperror.Wrap(apperror.ErrUpstream, "failed to fetch metadata", err)
	}

	if err := s.repo.Save(ctx, *metadata); err != nil {
		return apperror.Wrap(apperror.ErrInternal, "internal server error", err)
	}

	logger.WithRequest(ctx, zap.String("url", url)).Info(
		logger.Msg("Metadata", "Service", "Core", "metadata created successfully"),
	)
	return nil
}

// GetMetadata retrieves metadata for a URL, triggering background scraping if not found
func (s *metadataService) GetMetadata(ctx context.Context, url string) (*domain.Metadata, error) {
	if url == "" {
		return nil, apperror.New(apperror.ErrValidation, "url is required")
	}

	logger.WithRequest(ctx, zap.String("url", url)).Info(
		logger.Msg("Metadata", "Service", "Core", "getting metadata"),
	)

	metadata, err := s.repo.Get(ctx, url)
	if err == nil {
		logger.WithRequest(ctx, zap.String("url", url)).Info(
			logger.Msg("Metadata", "Service", "Core", "metadata found in database"),
		)
		return metadata, nil
	}

	if apperror.Is(err, apperror.ErrNotFound) {
		logger.WithRequest(ctx, zap.String("url", url)).Info(
			logger.Msg("Metadata", "Service", "Core", "metadata not found, queueing background scrape"),
		)

		select {
		case s.taskQueue <- scrapeTask{url: url, traceID: logger.ReqID(ctx)}:
			logger.WithRequest(ctx, zap.String("url", url)).Debug(
				logger.Msg("Metadata", "Service", "Core", "task queued successfully"),
			)
		default:
			logger.WithRequest(ctx, zap.String("url", url), zap.Int("queue_capacity", s.queueCapacity)).Warn(
				logger.Alert(logger.SeverityP2Medium, "Metadata", "Service", "Core", "task queue is full, rejecting request"),
			)
			return nil, apperror.New(apperror.ErrServiceUnavailable, "server busy, please retry later")
		}

		return nil, nil
	}

	return nil, apperror.Wrap(apperror.ErrInternal, "internal server error", err)
}

// Shutdown gracefully stops the worker pool
func (s *metadataService) Shutdown() {
	logger.WithRequest(context.Background()).Info(
		logger.Msg("Metadata", "Service", "Core", "shutting down metadata service"),
	)

	s.cancel()

	close(s.taskQueue)

	s.workerPool.Wait()

	logger.WithRequest(context.Background()).Info(
		logger.Msg("Metadata", "Service", "Core", "metadata service shutdown complete"),
	)
}
