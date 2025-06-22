package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds the complete application configuration
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Webhook   WebhookConfig   `mapstructure:"webhook"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
}

// AppConfig contains general application configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	Port        int    `mapstructure:"port"`
	APIKey      string `mapstructure:"api_key"`
	LogLevel    string `mapstructure:"log_level"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level       string `mapstructure:"level"`
	Environment string `mapstructure:"environment"`
	EnableJSON  bool   `mapstructure:"enable_json"`
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	Username          string        `mapstructure:"username"`
	Password          string        `mapstructure:"password"`
	Database          string        `mapstructure:"database"`
	SSLMode           string        `mapstructure:"ssl_mode"`
	MaxConnections    int32         `mapstructure:"max_connections"`
	MinConnections    int32         `mapstructure:"min_connections"`
	MaxConnLifetime   time.Duration `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime   time.Duration `mapstructure:"max_conn_idle_time"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`
}

// RedisConfig contains Redis cache configuration
type RedisConfig struct {
	Host     string        `mapstructure:"host"`
	Port     int           `mapstructure:"port"`
	Password string        `mapstructure:"password"`
	DB       int           `mapstructure:"db"`
	TTL      time.Duration `mapstructure:"ttl"`
}

// WebhookConfig contains webhook service configuration
type WebhookConfig struct {
	URL              string        `mapstructure:"url"`
	AuthToken        string        `mapstructure:"auth_token"`
	Timeout          time.Duration `mapstructure:"timeout"`
	MaxRetries       int           `mapstructure:"max_retries"`
	RetryBackoffBase time.Duration `mapstructure:"retry_backoff_base"`
}

// SchedulerConfig contains background processing configuration
type SchedulerConfig struct {
	Enabled   bool          `mapstructure:"enabled"`
	Interval  time.Duration `mapstructure:"interval"`
	BatchSize int           `mapstructure:"batch_size"`
}

// Load loads configuration from files and environment variables
func Load() (*Config, error) {
	// Set default values
	setDefaults()

	// Configuration file paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	// Read .env file if exists
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		// If .env doesn't exist, try config.yaml
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		if err := viper.ReadInConfig(); err != nil {
			// If no config file exists, continue with defaults and env vars
			fmt.Println("No config file found, using defaults and environment variables")
		}
	}

	// Enable environment variable support
	viper.AutomaticEnv()

	// Environment variable prefixes and mappings
	setupEnvBindings()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "Message Dispatcher")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.port", 8080)
	viper.SetDefault("app.api_key", "test-api-key-123")
	viper.SetDefault("app.log_level", "info")

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.environment", "development")
	viper.SetDefault("logging.enable_json", false)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.username", "msg_dispatcher_user")
	viper.SetDefault("database.password", "msg_dispatcher_pass123")
	viper.SetDefault("database.database", "message_dispatcher")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_connections", 25)
	viper.SetDefault("database.min_connections", 5)
	viper.SetDefault("database.max_conn_lifetime", "1h")
	viper.SetDefault("database.max_conn_idle_time", "30m")
	viper.SetDefault("database.health_check_period", "1m")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.ttl", "5m")

	// Webhook defaults
	viper.SetDefault("webhook.url", "https://webhook.site/a25c4f75-0f22-47f4-9def-dbdac00515ae")
	viper.SetDefault("webhook.auth_token", "")
	viper.SetDefault("webhook.timeout", "30s")
	viper.SetDefault("webhook.max_retries", 3)
	viper.SetDefault("webhook.retry_backoff_base", "100ms")

	// Scheduler defaults
	viper.SetDefault("scheduler.enabled", true)
	viper.SetDefault("scheduler.interval", "2m")
	viper.SetDefault("scheduler.batch_size", 2)
}

// setupEnvBindings binds environment variables to config keys
func setupEnvBindings() {
	// App env bindings
	viper.BindEnv("app.name", "APP_NAME")
	viper.BindEnv("app.version", "APP_VERSION")
	viper.BindEnv("app.environment", "APP_ENV")
	viper.BindEnv("app.port", "PORT")
	viper.BindEnv("app.api_key", "API_KEY")
	viper.BindEnv("app.log_level", "LOG_LEVEL")

	// Logging env bindings
	viper.BindEnv("logging.level", "LOG_LEVEL")
	viper.BindEnv("logging.environment", "LOG_ENV")
	viper.BindEnv("logging.enable_json", "LOG_JSON")

	// Database env bindings
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.username", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.database", "DB_NAME")
	viper.BindEnv("database.ssl_mode", "DB_SSL_MODE")
	viper.BindEnv("database.max_connections", "DB_MAX_CONNECTIONS")
	viper.BindEnv("database.min_connections", "DB_MIN_CONNECTIONS")
	viper.BindEnv("database.max_conn_lifetime", "DB_MAX_CONN_LIFETIME")
	viper.BindEnv("database.max_conn_idle_time", "DB_MAX_CONN_IDLE_TIME")
	viper.BindEnv("database.health_check_period", "DB_HEALTH_CHECK_PERIOD")

	// Redis env bindings
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")
	viper.BindEnv("redis.ttl", "REDIS_TTL")

	// Webhook env bindings
	viper.BindEnv("webhook.url", "WEBHOOK_URL")
	viper.BindEnv("webhook.auth_token", "WEBHOOK_AUTH_TOKEN")
	viper.BindEnv("webhook.timeout", "WEBHOOK_TIMEOUT")
	viper.BindEnv("webhook.max_retries", "WEBHOOK_MAX_RETRIES")
	viper.BindEnv("webhook.retry_backoff_base", "WEBHOOK_RETRY_BACKOFF_BASE")

	// Scheduler env bindings
	viper.BindEnv("scheduler.enabled", "SCHEDULER_ENABLED")
	viper.BindEnv("scheduler.interval", "SCHEDULER_INTERVAL")
	viper.BindEnv("scheduler.batch_size", "SCHEDULER_BATCH_SIZE")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.App.Port <= 0 {
		return fmt.Errorf("invalid port: %d", c.App.Port)
	}

	if c.App.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}

	if c.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}

	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	if c.Webhook.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	if c.Scheduler.BatchSize <= 0 {
		return fmt.Errorf("scheduler batch size must be positive")
	}

	if c.Scheduler.Interval <= 0 {
		return fmt.Errorf("scheduler interval must be positive")
	}

	return nil
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// BuildDatabaseConnectionString builds PostgreSQL connection string
func (c *Config) BuildDatabaseConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.Username,
		c.Database.Password, c.Database.Database, c.Database.SSLMode,
	)
}
