# Tests

This directory contains integration tests for the HTTP Metadata Inventory Service.

## Test Structure

```
tests/
└── integration/
    └── integration_test.go  # Integration tests with real MongoDB and HTTP
```

## Running Tests

### Unit Tests Only (Fast)
```bash
make test-unit
# or
go test -v -short ./...
```

### Integration Tests (Requires MongoDB)
```bash
# Start MongoDB first
docker run -d -p 27017:27017 --name test-mongo mongo:7

# Run integration tests
make test-integration
# or
go test -v -run Integration ./tests/integration/

# Cleanup
docker stop test-mongo && docker rm test-mongo
```

### All Tests
```bash
make test
# or
go test -v ./...
```

### Coverage Report
```bash
make test-coverage
```

## Integration Tests

Integration tests verify the full stack with real dependencies:

1. **TestIntegration_CreateAndGetMetadata**: End-to-end flow
2. **TestIntegration_GetMetadata_NotFound_BackgroundScrape**: Async worker behavior
3. **TestIntegration_Repository_SaveAndGet**: MongoDB persistence
4. **TestIntegration_Repository_Get_NotFound**: Error handling
5. **TestIntegration_Scraper_Fetch**: HTTP scraping

## Mocks

Generated mocks are in `/mocks/mock_ports.go` using `mockgen`:

```bash
make mock
```

## CI/CD

For CI pipelines, use:
```bash
# Fast unit tests
go test -v -short ./...

# Full integration tests (requires MongoDB)
docker-compose -f deployments/docker-compose.yml up -d mongo
go test -v ./tests/integration/
docker-compose -f deployments/docker-compose.yml down
```
