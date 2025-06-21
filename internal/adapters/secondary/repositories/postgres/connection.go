package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host              string
	Port              int
	Username          string
	Password          string
	Database          string
	SSLMode           string
	MaxConnections    int32
	MinConnections    int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// DefaultDatabaseConfig returns default database configuration
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:              "localhost",
		Port:              5432,
		Username:          "msg_dispatcher_user",
		Password:          "msg_dispatcher_pass123",
		Database:          "message_dispatcher",
		SSLMode:           "disable",
		MaxConnections:    25,
		MinConnections:    5,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   time.Minute * 30,
		HealthCheckPeriod: time.Minute,
	}
}

// BuildConnectionString builds PostgreSQL connection string from config
func (c DatabaseConfig) BuildConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode,
	)
}

// NewConnectionPool creates a new PostgreSQL connection pool
func NewConnectionPool(ctx context.Context, config DatabaseConfig) (*pgxpool.Pool, error) {
	connectionString := config.BuildConnectionString()

	poolConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool settings
	poolConfig.MaxConns = config.MaxConnections
	poolConfig.MinConns = config.MinConnections
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = config.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// CloseConnectionPool closes the database connection pool gracefully
func CloseConnectionPool(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}

// HealthCheck performs a health check on the database connection
func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("database pool is nil")
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// GetPoolStats returns connection pool statistics
func GetPoolStats(pool *pgxpool.Pool) *pgxpool.Stat {
	if pool == nil {
		return nil
	}
	return pool.Stat()
}
