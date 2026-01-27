package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort   string
	MongoURI     string
	DatabaseName string
	LogLevel     string
	IsProduction bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName: getEnv("DATABASE_NAME", "metadata_inventory"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		IsProduction: getEnv("GO_ENV", "development") == "production",
	}

	if cfg.MongoURI == "" {
		return nil, fmt.Errorf("MONGO_URI is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
