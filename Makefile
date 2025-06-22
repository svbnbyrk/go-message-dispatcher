# Message Dispatcher Makefile

# Variables
APP_NAME := message-dispatcher
VERSION := 1.0.0
DOCKER_IMAGE := $(APP_NAME):$(VERSION)
BINARY_NAME := bin/$(APP_NAME)

# Go parameters
GO_CMD := go
GO_BUILD := $(GO_CMD) build
GO_CLEAN := $(GO_CMD) clean
GO_TEST := $(GO_CMD) test
GO_GET := $(GO_CMD) get
GO_MOD := $(GO_CMD) mod

# Build targets
.PHONY: help build clean test test-unit test-integration test-coverage
.PHONY: dev dev-up dev-down migrate-up migrate-down migrate-create
.PHONY: docker-build docker-run docker-push
.PHONY: lint fmt vet security-scan deps-update

# Default target
help: ## Show this help message
	@echo "Message Dispatcher - Available Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
dev: ## Start development server
	@echo "🚀 Starting development server..."
	$(GO_CMD) run cmd/server/main.go

dev-up: ## Start development dependencies (Docker)
	@echo "🐳 Starting development dependencies..."
	docker-compose up -d postgres redis
	@echo "⏳ Waiting for services to be ready..."
	@sleep 5
	@echo "✅ Development dependencies started"

dev-down: ## Stop development dependencies
	@echo "🛑 Stopping development dependencies..."
	docker-compose down
	@echo "✅ Development dependencies stopped"

dev-reset: dev-down ## Reset development environment
	@echo "♻️  Resetting development environment..."
	docker-compose down -v
	$(MAKE) dev-up

# Build
build: ## Build the application
	@echo "🔨 Building application..."
	mkdir -p bin
	$(GO_BUILD) -o $(BINARY_NAME) cmd/server/main.go
	@echo "✅ Build completed: $(BINARY_NAME)"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	$(GO_CLEAN)
	rm -rf bin/
	@echo "✅ Clean completed"

# Testing
test: ## Run all tests
	@echo "🧪 Running all tests..."
	$(GO_TEST) -v ./...

test-unit: ## Run unit tests only
	@echo "🧪 Running unit tests..."
	$(GO_TEST) -v ./tests/unit/...

test-integration: ## Run integration tests (requires database)
	@echo "🧪 Running integration tests..."
	$(GO_TEST) -v ./tests/integration/...

test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	$(GO_TEST) -v -coverprofile=coverage.out ./...
	$(GO_CMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Database
migrate-up: ## Run database migrations
	@echo "⬆️  Running database migrations..."
	./deployments/scripts/setup-database.sh

migrate-down: ## Rollback database migrations
	@echo "⬇️  Rolling back database migrations..."
	@echo "⚠️  Manual rollback required - check migrations folder"

migrate-create: ## Create new migration file
	@read -p "Migration name: " name; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	mkdir -p internal/adapters/secondary/repositories/postgres/migrations; \
	touch internal/adapters/secondary/repositories/postgres/migrations/$${timestamp}_$${name}.up.sql; \
	touch internal/adapters/secondary/repositories/postgres/migrations/$${timestamp}_$${name}.down.sql; \
	echo "✅ Migration files created:"; \
	echo "  - $${timestamp}_$${name}.up.sql"; \
	echo "  - $${timestamp}_$${name}.down.sql"

# Docker
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "✅ Docker image built: $(DOCKER_IMAGE)"

docker-run: ## Run application in Docker
	@echo "🐳 Running application in Docker..."
	docker-compose up --build

docker-push: ## Push Docker image to registry
	@echo "🐳 Pushing Docker image..."
	docker push $(DOCKER_IMAGE)

# Code Quality
lint: ## Run linter
	@echo "🔍 Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt: ## Format code
	@echo "✨ Formatting code..."
	$(GO_CMD) fmt ./...
	@echo "✅ Code formatted"

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	$(GO_CMD) vet ./...
	@echo "✅ No issues found"

security-scan: ## Run security scan
	@echo "🔒 Running security scan..."
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	gosec ./...

# Dependencies
deps-download: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	$(GO_MOD) download

deps-tidy: ## Tidy dependencies
	@echo "📦 Tidying dependencies..."
	$(GO_MOD) tidy

deps-update: ## Update dependencies
	@echo "📦 Updating dependencies..."
	$(GO_GET) -u ./...
	$(GO_MOD) tidy

# Production
prod-deploy: ## Deploy to production (placeholder)
	@echo "🚀 Production deployment..."
	@echo "⚠️  Implement your production deployment steps here"

# Utilities
logs: ## Show application logs
	@echo "📜 Showing application logs..."
	docker-compose logs -f app

logs-db: ## Show database logs
	@echo "📜 Showing database logs..."
	docker-compose logs -f postgres

logs-redis: ## Show Redis logs
	@echo "📜 Showing Redis logs..."
	docker-compose logs -f redis

status: ## Show service status
	@echo "📊 Service status:"
	docker-compose ps

# Testing shortcuts
test-webhook: ## Run webhook test script
	@echo "🧪 Running webhook test..."
	./test_webhook.sh

test-background: ## Run background processing test
	@echo "🧪 Running background processing test..."
	./test_background_processing.sh

# Combined workflows
full-test: clean build test-unit test-integration ## Run complete test suite
	@echo "✅ Full test suite completed"

ci: deps-tidy fmt vet lint test ## Run CI pipeline
	@echo "✅ CI pipeline completed"

setup: deps-download dev-up migrate-up ## Initial project setup
	@echo "✅ Project setup completed"
	@echo "🎉 Ready to start development!"
	@echo ""
	@echo "Next steps:"
	@echo "  make dev          # Start development server"
	@echo "  make test         # Run tests"
	@echo "  make test-webhook # Test webhook integration" 