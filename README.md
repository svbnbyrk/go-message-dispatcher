# Go Message Dispatcher

An automatic message sending system built with Go that processes messages from a database queue and sends them to external webhook endpoints at regular intervals.

## Features

- 🔄 **Automatic Processing**: Sends messages every 2 minutes with configurable batch sizes
- 📊 **Status Tracking**: Complete message lifecycle management (PENDING → SENT → FAILED)
- 🔁 **Retry Logic**: Exponential backoff with configurable retry attempts
- 🏗️ **Clean Architecture**: Hexagonal architecture with clear separation of concerns
- 📊 **Caching**: Redis integration for performance optimization
- 🔍 **Observability**: Comprehensive logging and monitoring
- 🚀 **Production Ready**: Docker support with health checks

## Architecture

This project follows **Hexagonal Architecture** principles:

- **Domain Layer**: Pure business logic and entities
- **Application Layer**: Use cases and business workflows  
- **Infrastructure Layer**: Database, HTTP, external services
- **Dependency Injection**: Clean separation and testability

## Prerequisites

- Go 1.24.4 or higher
- Docker and Docker Compose
- PostgreSQL (via Docker)
- Redis (via Docker)

## Quick Start

### 1. Clone the repository
```bash
git clone https://github.com/svbnbyrk/go-message-dispatcher.git
cd go-message-dispatcher
```

### 2. Development Environment Setup
```bash
# Start local dependencies (PostgreSQL + Redis)
make dev-up

# Run database migrations
make migrate-up

# Install dependencies
go mod tidy
```

### 3. Environment Configuration
```bash
# Copy example environment file
cp .env.example .env

# Edit configuration as needed
vim .env
```

### 4. Run the application
```bash
# Development mode
make dev

# Or run directly
go run cmd/server/main.go
```

## Project Structure

```
go-message-dispatcher/
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── core/              # Domain and use cases (business logic)
│   ├── adapters/          # Infrastructure adapters
│   └── shared/            # Shared utilities
├── tests/                 # Test files
├── deployments/           # Docker and deployment configs
└── docs/                  # Documentation
```

## API Endpoints (Planned)

### Message Management
- `POST /api/v1/messages` - Create new message
- `GET /api/v1/messages` - List messages with pagination
- `GET /api/v1/messages/{id}` - Get message details

### Processing Control
- `POST /api/v1/messaging/start` - Start automatic processing
- `POST /api/v1/messaging/stop` - Stop automatic processing
- `GET /api/v1/messaging/status` - Get processor status

### System
- `GET /health` - Health check endpoint

## Development Commands

```bash
# Development
make dev              # Start development server
make dev-up           # Start development dependencies (Docker)
make dev-down         # Stop development dependencies

# Database
make migrate-up       # Run database migrations
make migrate-down     # Rollback database migrations
make migrate-create   # Create new migration file

# Testing
make test             # Run all tests
make test-unit        # Run unit tests only
make test-integration # Run integration tests
make test-coverage    # Run tests with coverage

# Code Quality
make lint             # Run linter
make fmt              # Format code
make vet              # Run go vet

# Build
make build            # Build application
make docker-build     # Build Docker image
```

## Configuration

Key environment variables:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=message_dispatcher
DB_USER=postgres
DB_PASSWORD=postgres

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Webhook
WEBHOOK_URL=https://your-webhook-endpoint.com
WEBHOOK_TIMEOUT=30s

# Processing
BATCH_SIZE=2
PROCESSING_INTERVAL=2m
MAX_RETRIES=3
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
