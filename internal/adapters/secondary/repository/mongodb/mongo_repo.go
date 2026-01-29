package mongodb

import (
	"context"
	"time"

	"github.com/codevshl/http-metadata-inventory-service/internal/core/apperror"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/domain"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/ports"
	"github.com/codevshl/http-metadata-inventory-service/internal/platform/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const collectionName = "metadata"

type mongoRepository struct {
	client     *mongo.Client
	dbName     string
	collection *mongo.Collection
}

// NewMongoRepository creates a new MongoDB repository and ensures indexes
func NewMongoRepository(client *mongo.Client, dbName string) ports.MetadataRepository {
	repo := &mongoRepository{
		client:     client,
		dbName:     dbName,
		collection: client.Database(dbName).Collection(collectionName),
	}

	if err := repo.ensureIndexes(context.Background()); err != nil {
		logger.WithRequest(context.Background(), zap.Error(err)).Error(
			logger.Alert(logger.SeverityP1High, "Metadata", "MongoRepository", "Adapter", "failed to create indexes"),
		)
	}

	return repo
}

// ensureIndexes creates necessary indexes for the collection
func (r *mongoRepository) ensureIndexes(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "url", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return apperror.Wrap(apperror.ErrInternal, "internal server error", err)
	}

	logger.WithRequest(ctx).Info(
		logger.Msg("Metadata", "MongoRepository", "Adapter", "mongodb indexes created successfully"),
	)
	return nil
}

// Save stores or updates metadata in the database
func (r *mongoRepository) Save(ctx context.Context, metadata domain.Metadata) error {
	mongoModel := toMongoModel(metadata)

	filter := bson.M{"url": metadata.URL}
	opts := options.Replace().SetUpsert(true)

	_, err := r.collection.ReplaceOne(ctx, filter, mongoModel, opts)
	if err != nil {
		return apperror.Wrap(apperror.ErrInternal, "internal server error", err)
	}

	return nil
}

// Get retrieves metadata for a given URL
func (r *mongoRepository) Get(ctx context.Context, url string) (*domain.Metadata, error) {
	filter := bson.M{"url": url}

	var mongoModel mongoMetadata
	err := r.collection.FindOne(ctx, filter).Decode(&mongoModel)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, apperror.New(apperror.ErrNotFound, "metadata not found")
		}
		return nil, apperror.Wrap(apperror.ErrInternal, "internal server error", err)
	}

	domainModel := toDomain(mongoModel)
	return &domainModel, nil
}
