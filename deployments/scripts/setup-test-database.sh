#!/bin/bash

# Test Database Setup Script for Message Dispatcher Application
# This script sets up PostgreSQL test database for integration tests

set -e

# Configuration
CONTAINER_NAME="msg-dispatcher-test-db"
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="message_dispatcher_test"
DB_USER="msg_dispatcher_user"
DB_PASSWORD="msg_test_pass123"
POSTGRES_PASSWORD="postgres"

echo "🧪 Setting up Test Database for Message Dispatcher..."

# Check if container already exists
if docker ps -a | grep -q $CONTAINER_NAME; then
    echo "🔄 Container '$CONTAINER_NAME' already exists. Stopping and removing..."
    docker stop $CONTAINER_NAME 2>/dev/null || true
    docker rm $CONTAINER_NAME 2>/dev/null || true
fi

# Start PostgreSQL container
echo "🐳 Starting PostgreSQL container..."
docker run --name $CONTAINER_NAME \
    -p $DB_PORT:5432 \
    -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
    -e POSTGRES_DB=postgres \
    -d postgres:15

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 5

# Retry connection check
for i in {1..10}; do
    if docker exec $CONTAINER_NAME pg_isready -U postgres > /dev/null 2>&1; then
        echo "✅ PostgreSQL is ready"
        break
    fi
    echo "   Attempt $i/10: PostgreSQL not ready yet, waiting..."
    sleep 2
done

# Create test database
echo "📦 Creating test database '$DB_NAME'..."
docker exec $CONTAINER_NAME psql -U postgres -c "CREATE DATABASE $DB_NAME;" 2>/dev/null || echo "Database '$DB_NAME' already exists"

# Create test user
echo "👤 Creating test user '$DB_USER'..."
docker exec $CONTAINER_NAME psql -U postgres -c "CREATE USER $DB_USER WITH ENCRYPTED PASSWORD '$DB_PASSWORD';" 2>/dev/null || echo "User '$DB_USER' already exists"

# Grant privileges
echo "🔐 Granting privileges..."
docker exec $CONTAINER_NAME psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"
docker exec $CONTAINER_NAME psql -U postgres -d $DB_NAME -c "GRANT ALL ON SCHEMA public TO $DB_USER;"
docker exec $CONTAINER_NAME psql -U postgres -d $DB_NAME -c "GRANT ALL ON ALL TABLES IN SCHEMA public TO $DB_USER;"
docker exec $CONTAINER_NAME psql -U postgres -d $DB_NAME -c "GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;"
docker exec $CONTAINER_NAME psql -U postgres -d $DB_NAME -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;"

# Run migrations
echo "🔄 Running database migrations..."
MIGRATION_FILE="internal/adapters/secondary/repositories/postgres/migrations/001_create_messages.up.sql"

if [ -f "$MIGRATION_FILE" ]; then
    docker exec -i $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME < "$MIGRATION_FILE"
    echo "✅ Migrations applied successfully"
else
    echo "⚠️  Migration file not found: $MIGRATION_FILE"
    echo "   Please ensure you're running this script from the project root"
fi

# Test connection
echo "🧪 Testing connection..."
if docker exec $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME -c '\q' 2>/dev/null; then
    echo "✅ Test database connection successful"
else
    echo "❌ Error: Cannot connect to test database"
    exit 1
fi

echo ""
echo "🎉 Test database setup completed successfully!"
echo ""
echo "Test Database Configuration:"
echo "  Container: $CONTAINER_NAME"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo "  Password: $DB_PASSWORD"
echo ""
echo "You can now run integration tests with:"
echo "  go test ./tests/integration/ -v"
echo ""
echo "To stop and cleanup the test database:"
echo "  docker stop $CONTAINER_NAME && docker rm $CONTAINER_NAME" 