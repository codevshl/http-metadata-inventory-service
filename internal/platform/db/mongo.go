package db

import (
	"context"
	"sync"
	"time"

	"github.com/codevshl/http-metadata-inventory-service/internal/platform/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

var (
	client     *mongo.Client
	clientOnce sync.Once
)

// Connect initializes the MongoDB client singleton
func Connect(uri string) error {
	var err error
	clientOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOptions := options.Client().ApplyURI(uri)
		// TODO: we're using default connection pool settings here. for high load,
		// we should probably tweak MaxPoolSize and MaxConnIdleTime to avoid
		// keeping too many connections open or running out of them.
		client, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			logger.WithRequest(context.Background(), zap.Error(err)).Error(
				logger.Alert(logger.SeverityP1High, "Database", "Mongo", "Platform", "failed to create mongo client"),
			)
			return
		}

		// Ping the database
		if err = client.Ping(ctx, readpref.Primary()); err != nil {
			logger.WithRequest(context.Background(), zap.Error(err)).Error(
				logger.Alert(logger.SeverityP1High, "Database", "Mongo", "Platform", "failed to ping mongo"),
			)
			return
		}

		logger.WithRequest(context.Background()).Info(
			logger.Msg("Database", "Mongo", "Platform", "successfully connected to MongoDB"),
		)
	})

	return err
}

// GetClient returns the MongoDB client instance
func GetClient() *mongo.Client {
	if client == nil {
		panic("MongoDB client is not initialized. Call db.Connect() first.")
	}
	return client
}

// Disconnect closes the MongoDB connection
func Disconnect() {
	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			logger.WithRequest(context.Background(), zap.Error(err)).Error(
				logger.Alert(logger.SeverityP2Medium, "Database", "Mongo", "Platform", "failed to disconnect from MongoDB"),
			)
		} else {
			logger.WithRequest(context.Background()).Info(
				logger.Msg("Database", "Mongo", "Platform", "disconnected from MongoDB"),
			)
		}
	}
}
