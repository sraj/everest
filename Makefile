.PHONY: build run test clean dev docker-up docker-down lint fmt tidy migrate-up migrate-down migrate-version seed setup fresh web-install web-dev web-build

# Build variables
BINARY_NAME=everest
MIGRATE_BINARY=migrate
BUILD_DIR=./bin
CMD_DIR=./cmd/server
MIGRATE_DIR=./cmd/migrate

# Version variables (embedded via ldflags)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown")
LDFLAGS = -X github.com/sraj/everest/internal/version.Version=$(VERSION) -X github.com/sraj/everest/internal/version.Commit=$(COMMIT) -X github.com/sraj/everest/internal/version.BuildDate=$(BUILD_DATE)

# Go variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

# Build the application
build:
	@echo "Building server..."
	@go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

# Build the migrate CLI
build-migrate:
	@echo "Building migrate CLI..."
	@go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(MIGRATE_BINARY) $(MIGRATE_DIR)

# Generate protobuf stubs and OpenAPI spec
proto:
	@echo "Generating protobuf stubs + OpenAPI spec..."
	@cd api && buf dep update && buf generate
	@echo "  Go stubs:  api/gen/go/"
	@echo "  OpenAPI:   api/gen/openapiv2/api.swagger.json"

proto-lint:
	@echo "Linting protobuf files..."
	@cd api && buf lint

# Run the application
run: build
	@echo "Running..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Run with air for hot reloading
dev:
	@echo "Starting development server with hot reload..."
	@air

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Start all services in Docker (full stack)
docker-up:
	@echo "Starting all Docker services..."
	@docker compose up -d --build

# Start infrastructure services only (for local backend/frontend dev)
docker-infra:
	@echo "Starting infrastructure services (PostgreSQL, MinIO, Bifrost)..."
	@docker compose up -d --build postgres minio chromadb bifrost

# Stop docker services
docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/bufbuild/buf/cmd/buf@latest

# Database migrations
migrate-up: build-migrate
	@echo "Running migrations..."
	@$(BUILD_DIR)/$(MIGRATE_BINARY) up

migrate-down: build-migrate
	@echo "Rolling back migrations..."
	@$(BUILD_DIR)/$(MIGRATE_BINARY) down

migrate-version: build-migrate
	@$(BUILD_DIR)/$(MIGRATE_BINARY) version

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir db/migrations -seq $$name

# Database seeding
seed: build-migrate
	@echo "Running seeds..."
	@$(BUILD_DIR)/$(MIGRATE_BINARY) seed

seed-status: build-migrate
	@$(BUILD_DIR)/$(MIGRATE_BINARY) seed:status

seed-reset: build-migrate
	@$(BUILD_DIR)/$(MIGRATE_BINARY) seed:reset

# Combined database commands
setup: build-migrate
	@echo "Running migrations and seeds..."
	@$(BUILD_DIR)/$(MIGRATE_BINARY) setup

fresh: build-migrate
	@echo "Fresh database (drop, migrate, seed)..."
	@$(BUILD_DIR)/$(MIGRATE_BINARY) fresh

# Frontend commands
web-install:
	@echo "Installing frontend dependencies..."
	@cd web && npm install

web-dev:
	@echo "Starting frontend dev server..."
	@cd web && npm run dev

web-build:
	@echo "Building frontend..."
	@cd web && npm run build

# Help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Backend:"
	@echo "  make build           - Build the server application"
	@echo "  make build-migrate   - Build the migrate CLI"
	@echo "  make run             - Build and run the application"
	@echo "  make dev             - Run with hot reload (requires air)"
	@echo "  make test            - Run tests"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make lint            - Run linter"
	@echo "  make fmt             - Format code"
	@echo "  make tidy            - Tidy dependencies"
	@echo "  make proto           - Generate protobuf stubs + OpenAPI spec"
	@echo ""
	@echo "Database:"
	@echo "  make docker-up       - Start Docker services (PostgreSQL, MinIO)"
	@echo "  make docker-down     - Stop Docker services"
	@echo "  make migrate-up      - Apply all pending migrations"
	@echo "  make migrate-down    - Roll back all migrations"
	@echo "  make migrate-version - Show current migration version"
	@echo "  make migrate-create  - Create a new migration file"
	@echo "  make seed            - Run all pending seeds"
	@echo "  make seed-status     - Show seed status"
	@echo "  make seed-reset      - Reset seed tracking"
	@echo "  make setup           - Run migrations + seeds"
	@echo "  make fresh           - Drop, migrate, and seed (fresh start)"
	@echo ""
	@echo "Frontend:"
	@echo "  make web-install     - Install frontend dependencies"
	@echo "  make web-dev         - Start frontend dev server"
	@echo "  make web-build       - Build frontend for production"
	@echo ""
	@echo "Other:"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make install-tools   - Install development tools"
