package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort     string
	MongoURI       string
	DatabaseName   string
	LogLevel       string
	IsProduction   bool
	WorkerPoolSize int
	TaskQueueSize  int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Attempt to load .env file, but don't fail if it doesn't exist (e.g. in prod)
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName:   getEnv("DATABASE_NAME", "metadata_inventory"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		IsProduction:   getEnv("GO_ENV", "development") == "production",
		WorkerPoolSize: getEnvInt("WORKER_POOL_SIZE", 10),
		TaskQueueSize:  getEnvInt("TASK_QUEUE_SIZE", 100),
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

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}
