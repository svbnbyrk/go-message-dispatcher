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
	RetryBackoffMax  time.Duration `mapstructure:"retry_backoff_max"`
}

// SchedulerConfig contains background processing configuration
type SchedulerConfig struct {
	Enabled   bool          `mapstructure:"enabled"`
	Interval  time.Duration `mapstructure:"interval"`
	BatchSize int           `mapstructure:"batch_size"`
}

// Load loads configuration from config.yaml file only
func Load() (*Config, error) {
	// Set configuration file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../../")
	viper.AddConfigPath("./configs")

	// Read config.yaml file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
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
