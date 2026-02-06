// @title HTTP Metadata Inventory Service API
// @version 1.1.0
// @description Collects and serves HTTP metadata for a given URL.
// @description
// @description **How it works**
// @description - POST /api/v1/metadata performs an immediate scrape and persists the result.
// @description - GET /api/v1/metadata returns cached metadata if present.
// @description - If not present, GET returns 202 Accepted and starts a background scrape.
// @BasePath /api/v1
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	adapter_http "github.com/codevshl/http-metadata-inventory-service/internal/adapters/primary/http"
	"github.com/codevshl/http-metadata-inventory-service/internal/adapters/secondary/repository/mongodb"
	"github.com/codevshl/http-metadata-inventory-service/internal/adapters/secondary/scraper/http_client"
	"github.com/codevshl/http-metadata-inventory-service/internal/config"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/ports"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/services"
	"github.com/codevshl/http-metadata-inventory-service/internal/platform/db"
	"github.com/codevshl/http-metadata-inventory-service/internal/platform/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := loadConfig()

	initLogger(cfg)
	defer logger.Sync()
	defer func() {
		if r := recover(); r != nil {
			logger.Get().Fatal("fatal panic", zap.Any("panic", r), zap.Stack("stack"))
		}
	}()

	initDatabase(cfg)
	defer db.Disconnect()

	svc := buildService(cfg)
	server := buildServer(cfg, svc)

	startServer(server)
	waitForShutdown(server, svc)

	logger.Info("Server exited gracefully")
}

func loadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		panic("invalid configuration: " + err.Error())
	}
	return cfg
}

func initLogger(cfg *config.Config) {
	logger.Init(cfg.LogLevel, cfg.IsProduction)

	logger.Info(
		"Starting HTTP Metadata Inventory Service",
		zap.String("port", cfg.ServerPort),
		zap.Int("workers", cfg.WorkerPoolSize),
	)
}

func initDatabase(cfg *config.Config) {
	if err := db.Connect(cfg.MongoURI); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
}

func buildService(cfg *config.Config) ports.MetadataService {
	repo := mongodb.NewMongoRepository(
		db.GetClient(),
		cfg.DatabaseName,
	)

	scraper := http_client.NewScraper()

	svcConfig := services.Config{
		WorkerPoolSize: cfg.WorkerPoolSize,
		TaskQueueSize:  cfg.TaskQueueSize,
	}

	return services.NewMetadataService(repo, scraper, svcConfig)
}

func buildServer(cfg *config.Config, svc ports.MetadataService) *http.Server {
	router := adapter_http.SetupRouter(svc, cfg.IsProduction)

	return &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func startServer(srv *http.Server) {
	go func() {
		logger.Info("Server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()
}

func waitForShutdown(srv *http.Server, svc ports.MetadataService) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	if shutdownable, ok := svc.(interface{ Shutdown() }); ok {
		shutdownable.Shutdown()
	}
}
