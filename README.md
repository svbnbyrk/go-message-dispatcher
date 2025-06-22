# Go Message Dispatcher

An automatic message sending system built with Go that processes messages from a database queue and sends them to external webhook endpoints at regular intervals.

## Features

- 🔄 **Automatic Processing**: Sends messages every 2 minutes with configurable batch sizes
- 📊 **Status Tracking**: Complete message lifecycle management (PENDING → SENT → FAILED)
- 🔁 **Retry Logic**: Exponential backoff with configurable retry attempts
- 🏗️ **Clean Architecture**: Hexagonal architecture with clear separation of concerns
- 📊 **Caching**: Redis integration for performance optimization
- 🔍 **Observability**: Comprehensive logging and monitoring
- ⚡ **Distributed Safe**: Race condition prevention with PostgreSQL row locking
- 🚀 **Production Ready**: Docker support with health checks and scalability

## Architecture

This project follows **Hexagonal Architecture** principles:

- **Domain Layer**: Pure business logic and entities
- **Application Layer**: Use cases and business workflows  
- **Infrastructure Layer**: Database, HTTP, external services
- **Dependency Injection**: Clean separation and testability

## Prerequisites

- Go 1.24+ or higher
- Docker and Docker Compose
- Make (for using Makefile commands)

## Quick Start

### 1. Clone and Setup
```bash
git clone https://github.com/svbnbyrk/go-message-dispatcher.git
cd go-message-dispatcher
```

### 2. Start Development Environment
```bash
# Start all dependencies (PostgreSQL + Redis) and run the application
make dev-up && make dev
```

That's it! The application will be running at `http://localhost:8080`

### Configuration

Before running, update the `config.yaml` file with your settings:

```yaml
# Key configurations to update:
app:
  api_key: "your-secure-api-key-here"  # Change this!

webhook:
  url: "https://your-webhook-endpoint.com"  # Your webhook URL
  auth_token: "your-webhook-auth-token"     # Optional webhook auth

database:
  # Database settings (defaults work for development)
  password: "msg_dispatcher_pass123"

redis:
  # Redis settings (defaults work for development)
  password: ""
```

### API Access

Once running, you can access:
- **API Base**: `http://localhost:8080`
- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **Health Check**: `http://localhost:8080/health`

**Authentication**: All API endpoints (except `/health`) require Bearer token:
```bash
curl -H "Authorization: Bearer your-secure-api-key-here" \
     http://localhost:8080/api/v1/messages
```

## Project Structure

```
go-message-dispatcher/
├── cmd/                   # Application entry points
├── internal/              # Private application code
│   ├── core/              # Domain and use cases (business logic)
│   ├── adapters/          # Infrastructure adapters
│   └── shared/            # Shared utilities
├── tests/                 # Test files
├── deployments/           # Docker and deployment configs
├── docs/                  # Documentation
└── config.yaml            # Main configuration file
```

## API Documentation

### Swagger UI
The API is fully documented with Swagger/OpenAPI 3.0. Once the server is running, you can access:

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **OpenAPI JSON**: http://localhost:8080/swagger/doc.json

### Quick API Reference

#### Message Management
- `POST /api/v1/messages` - Create new message
- `GET /api/v1/messages` - List messages with pagination
- `GET /api/v1/messages/{id}` - Get message details

#### Message Processing
- `POST /api/v1/messaging/process` - Manually trigger processing
- `GET /api/v1/messaging/status` - Get processing status

#### Scheduler Management
- `POST /api/v1/scheduler/start` - Start background scheduler
- `POST /api/v1/scheduler/stop` - Stop background scheduler
- `GET /api/v1/scheduler/status` - Get scheduler status

#### System
- `GET /health` - Health check endpoint

## Development Commands

```bash
# Quick Start
make dev-up           # Start dependencies (PostgreSQL + Redis)
make dev              # Start development server

# Development
make dev-down         # Stop development dependencies

# Database
make migrate-up       # Run database migrations
make migrate-down     # Rollback database migrations
make migrate-create   # Create new migration file

# Documentation
make swagger          # Generate and serve swagger docs
make swagger-gen      # Generate swagger documentation
make swagger-serve    # Generate docs and start server

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

## Example Usage

### 1. Create a Message
```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer your-secure-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "phoneNumber": "+905551234567",
    "content": "Hello, this is a test message!"
  }'
```

### 2. List Messages
```bash
curl -H "Authorization: Bearer your-secure-api-key-here" \
     "http://localhost:8080/api/v1/messages?limit=10&offset=0"
```

### 3. Check Processing Status
```bash
curl -H "Authorization: Bearer your-secure-api-key-here" \
     http://localhost:8080/api/v1/messaging/status
```

## Distributed Deployment

### Race Condition Prevention
This system is designed to run safely with multiple instances:

- ✅ **PostgreSQL Row Locking**: Uses `FOR UPDATE SKIP LOCKED` to prevent duplicate message processing
- ✅ **Atomic Operations**: Each instance locks different message batches atomically  
- ✅ **No Shared State**: Stateless design allows horizontal scaling
- ✅ **Graceful Degradation**: If one instance fails, others continue processing

### Scaling Guidelines
```bash
# Multiple instances can run simultaneously
docker-compose up --scale app=3

# Each instance will:
# 1. Process different message batches
# 2. Never duplicate webhook calls  
# 3. Handle backlog efficiently
```

### Testing Distributed Behavior
```bash
# Test race condition prevention
make test-race-condition

# Test deployment scenario with existing pending messages
make test-deployment
```

## Configuration Reference

All configuration is managed through `config.yaml`. Key settings:

```yaml
app:
  api_key: "your-secure-api-key-here"  # API authentication key
  port: 8080                           # Server port

database:
  host: "localhost"                    # Database host
  port: 5432                          # Database port
  username: "msg_dispatcher_user"      # Database username
  password: "msg_dispatcher_pass123"   # Database password
  database: "message_dispatcher"       # Database name

redis:
  host: "localhost"                    # Redis host
  port: 6379                          # Redis port
  password: ""                        # Redis password (empty for no auth)

webhook:
  url: "https://your-webhook-endpoint.com"  # Target webhook URL
  auth_token: ""                           # Optional webhook authentication
  timeout: "30s"                           # Request timeout
  max_retries: 3                           # Retry attempts

scheduler:
  enabled: true                        # Enable automatic processing
  interval: "2m"                       # Processing interval
  batch_size: 2                        # Messages per batch
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
