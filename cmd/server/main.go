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
	"github.com/codevshl/http-metadata-inventory-service/internal/core/services"
	"github.com/codevshl/http-metadata-inventory-service/internal/platform/db"
	"github.com/codevshl/http-metadata-inventory-service/internal/platform/logger"
	"go.uber.org/zap"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		panic("invalid configuration: " + err.Error())
	}

	// 2. Initialize Logger (Singleton)
	logger.Init(cfg.LogLevel, cfg.IsProduction)
	defer logger.Sync()

	logger.Info("Starting HTTP Metadata Inventory Service", zap.String("port", cfg.ServerPort))

	// 3. Initialize Database (Singleton)
	if err := db.Connect(cfg.MongoURI); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Disconnect()

	repo := mongodb.NewMongoRepository(db.GetClient(), cfg.DatabaseName)

	scraper := http_client.NewScraper()

	// 5. Initialize Services (Core)
	svc := services.NewMetadataService(repo, scraper)

	// 6. Initialize Router (Driving/Primary)
	router := adapter_http.SetupRouter(svc, cfg.IsProduction)

	// 7. Start Server (Graceful Shutdown)
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Run server in a goroutine so it doesn't block shutdown handling
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Listen: %s\n", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}
