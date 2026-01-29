# HTTP Metadata Inventory Service

A production-grade Go microservice that collects and stores HTTP metadata (headers, cookies, page source) for URLs. Built with hexagonal architecture, featuring async background workers and MongoDB persistence.

## Features

- 🚀 **Async Processing**: Background worker pool for non-blocking metadata collection
- 🏗️ **Hexagonal Architecture**: Clean separation of domain, ports, and adapters
- 🔄 **Graceful Shutdown**: Proper resource cleanup and worker draining
- 📊 **MongoDB Storage**: Persistent storage with automatic indexing
- 🐳 **Docker Ready**: One-command deployment with Docker Compose
- 📝 **Structured Logging**: Production-ready logging with Zap

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local development)

### Run with Docker Compose

```bash
cd deployments
docker-compose up --build
```

The service will be available at `http://localhost:8080`

### Run Locally

1. Start MongoDB:
```bash
docker run -d -p 27017:27017 mongo:7
```

2. Run the service:
```bash
go run cmd/server/main.go
```

## API Documentation
OpenAPI 2.0 (Swagger) documentation is available at:
`http://localhost:8080/docs/index.html`

## API Endpoints

### POST /api/v1/metadata

Create metadata for a URL (synchronous scraping).

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/metadata \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

**Response:** `201 Created`

### GET /api/v1/metadata?url=<URL>

Retrieve metadata for a URL.

**Request:**
```bash
curl "http://localhost:8080/api/v1/metadata?url=https://example.com"
```

**Responses:**
- `200 OK` - Metadata found and returned
- `202 Accepted` - Metadata not found, background scraping initiated
- `400 Bad Request` - Invalid URL
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Server busy, please retry later

**Example Response (200):**
```json
{
  "url": "https://example.com",
  "headers": {
    "Content-Type": "text/html; charset=UTF-8",
    "Server": "nginx"
  },
  "cookies": {
    "session": "abc123"
  },
  "page_source": "<!DOCTYPE html>..."
}
```

### GET /health

Health check endpoint.

**Response:** `200 OK` with `{"status": "ok"}`

## Configuration

Configure via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP server port |
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB connection string |
| `DATABASE_NAME` | `metadata_inventory` | MongoDB database name |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `GO_ENV` | `development` | Environment (development, production) |
| `WORKER_POOL_SIZE` | `10` | Number of background workers |
| `TASK_QUEUE_SIZE` | `100` | Background task queue capacity |

Create a `.env` file (see `.env.example`):
```bash
cp .env.example .env
```

## Architecture

### Hexagonal Architecture (Ports & Adapters)

```
internal/
├── core/                    # Business Logic (The Hexagon)
│   ├── domain/             # Entities & Errors
│   ├── ports/              # Interfaces
│   └── services/           # Use Cases
├── adapters/
│   ├── primary/            # Driving Adapters (HTTP)
│   └── secondary/          # Driven Adapters (DB, Scraper)
├── config/                 # Configuration
└── platform/               # Infrastructure (Logger, DB)
```

### Background Worker Pattern

- **Worker Pool**: Fixed number of goroutines (default: 10)
- **Task Queue**: Buffered channel (default capacity: 100)
- **Non-blocking**: GET endpoint never blocks, even if queue is full
- **Graceful Shutdown**: Workers drain tasks before exit

## Development

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Run with Coverage

```bash
make test-coverage
```

### Run Unit Tests Only

```bash
make test-unit
```

### Run Integration Tests

Requires MongoDB running:

```bash
# Start MongoDB
docker run -d -p 27017:27017 --name test-mongo mongo:7

# Run integration tests
make test-integration

# Cleanup
docker stop test-mongo && docker rm test-mongo
```

### Lint

```bash
make lint
```

### Generate Mocks

```bash
make mock
```

## Project Structure

```
.
├── cmd/server/              # Application entry point
├── internal/                # Private application code
├── mocks/                   # Generated mocks
├── api/                     # API documentation
├── deployments/             # Docker & deployment configs
│   ├── Dockerfile
│   └── docker-compose.yml
├── scripts/                 # Utility scripts
├── Makefile                 # Build commands
└── README.md
```

## Design Decisions

### Why Worker Pool?
- **Resource Control**: Prevents goroutine explosion under load
- **Backpressure**: Queue provides natural rate limiting
- **Predictable**: Fixed resource usage

### Why Upsert in Repository?
- **Idempotency**: Safe to retry saves
- **Simplicity**: Single method for create/update

### Why 202 Accepted?
- **Responsiveness**: API never blocks on slow scrapes
- **User Experience**: Immediate feedback
- **Scalability**: Decouples request handling from processing

## Production Considerations

### Monitoring (Future)
- Metrics: queue depth, worker utilization, scrape latency
- Tracing: OpenTelemetry integration
- Alerts: SLO-based alerting

### Scaling
- Current: Single-process worker pool
- Future: Replace with Redis/RabbitMQ for multi-instance deployment

### Security
- TLS in production
- Rate limiting per client
- Input validation and sanitization

## License

MIT
