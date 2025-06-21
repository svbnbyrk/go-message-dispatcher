#!/bin/bash

# Database Setup Script for Message Dispatcher Application
# This script sets up PostgreSQL database, user, and schema for production

set -e

# Configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-message_dispatcher}
DB_USER=${DB_USER:-msg_dispatcher_user}
DB_PASSWORD=${DB_PASSWORD:-msg_dispatcher_pass123}
POSTGRES_USER=${POSTGRES_USER:-postgres}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-postgres}

echo "🚀 Setting up Message Dispatcher Database..."

# Check if PostgreSQL is running
echo "📊 Checking PostgreSQL connection..."
if ! PGPASSWORD=$POSTGRES_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $POSTGRES_USER -c '\q' 2>/dev/null; then
    echo "❌ Error: Cannot connect to PostgreSQL server"
    echo "   Make sure PostgreSQL is running on $DB_HOST:$DB_PORT"
    exit 1
fi

echo "✅ PostgreSQL connection successful"

# Create database
echo "📦 Creating database '$DB_NAME'..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $POSTGRES_USER -c "CREATE DATABASE $DB_NAME;" 2>/dev/null || echo "Database '$DB_NAME' already exists"

# Create user
echo "👤 Creating user '$DB_USER'..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $POSTGRES_USER -c "CREATE USER $DB_USER WITH ENCRYPTED PASSWORD '$DB_PASSWORD';" 2>/dev/null || echo "User '$DB_USER' already exists"

# Grant privileges
echo "🔐 Granting privileges..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $POSTGRES_USER -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"
PGPASSWORD=$POSTGRES_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $POSTGRES_USER -d $DB_NAME -c "GRANT ALL ON SCHEMA public TO $DB_USER;"
PGPASSWORD=$POSTGRES_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $POSTGRES_USER -d $DB_NAME -c "GRANT ALL ON ALL TABLES IN SCHEMA public TO $DB_USER;"
PGPASSWORD=$POSTGRES_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $POSTGRES_USER -d $DB_NAME -c "GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;"
PGPASSWORD=$POSTGRES_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $POSTGRES_USER -d $DB_NAME -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;"

# Run migrations
echo "🔄 Running database migrations..."
MIGRATION_FILE="../../internal/adapters/secondary/repositories/postgres/migrations/001_create_messages.up.sql"

if [ -f "$MIGRATION_FILE" ]; then
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$MIGRATION_FILE"
    echo "✅ Migrations applied successfully"
else
    echo "⚠️  Migration file not found: $MIGRATION_FILE"
    echo "   Please run migrations manually"
fi

# Test connection with application user
echo "🧪 Testing application user connection..."
if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c '\q' 2>/dev/null; then
    echo "✅ Application user connection successful"
else
    echo "❌ Error: Application user cannot connect to database"
    exit 1
fi

echo ""
echo "🎉 Database setup completed successfully!"
echo ""
echo "Database Configuration:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo "  Password: [HIDDEN]"
echo ""
echo "You can now start the Message Dispatcher application." 