.PHONY: run test test-unit test-integration test-coverage build docker-up docker-down clean mock lint

# Run locally (requires MongoDB running)
run:
	go run cmd/server/main.go

# Run all tests
test:
	go test -v ./...

# Run unit tests only
test-unit:
	go test -v -short ./...

# Run integration tests (requires MongoDB)
test-integration:
	go test -v -run Integration ./tests/integration/

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Build binary
build:
	go build -o bin/server cmd/server/main.go

# Start services with Docker Compose
docker-up:
	cd deployments && docker-compose up --build

# Stop Docker Compose services
docker-down:
	cd deployments && docker-compose down

# Clean up
clean:
	rm -f bin/server coverage.out
	cd deployments && docker-compose down -v

# Generate mocks using mockgen
mock:
	go install go.uber.org/mock/mockgen@latest
	mockgen -source=internal/core/ports/secondary.go -destination=mocks/mock_ports.go -package=mocks

# Lint code
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Tidy dependencies
tidy:
	go mod tidy
