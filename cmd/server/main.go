package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/primary/http/handlers"
	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/primary/http/routes"
	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/secondary/repositories/postgres"
	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/secondary/services/cache"
	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/secondary/services/webhook"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/services"
	schedulerServices "github.com/svbnbyrk/go-message-dispatcher/internal/core/services"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/usecases"
	"github.com/svbnbyrk/go-message-dispatcher/internal/shared/config"
	"github.com/svbnbyrk/go-message-dispatcher/internal/shared/logger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration first
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger with config
	loggerConfig := logger.LoggerConfig{
		Level:       cfg.Logging.Level,
		Environment: cfg.Logging.Environment,
		EnableJSON:  cfg.Logging.EnableJSON,
	}

	if err := logger.Initialize(loggerConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// Log system information
	logger.LogSystemInfo(cfg.App.Name, cfg.App.Version, cfg.App.Environment)

	// Initialize database
	logger.Info("Initializing database connection", zap.String("host", cfg.Database.Host))
	dbConfig := postgres.DatabaseConfig{
		Host:              cfg.Database.Host,
		Port:              cfg.Database.Port,
		Username:          cfg.Database.Username,
		Password:          cfg.Database.Password,
		Database:          cfg.Database.Database,
		SSLMode:           cfg.Database.SSLMode,
		MaxConnections:    cfg.Database.MaxConnections,
		MinConnections:    cfg.Database.MinConnections,
		MaxConnLifetime:   cfg.Database.MaxConnLifetime,
		MaxConnIdleTime:   cfg.Database.MaxConnIdleTime,
		HealthCheckPeriod: cfg.Database.HealthCheckPeriod,
	}

	ctx := context.Background()
	pool, err := postgres.NewConnectionPool(ctx, dbConfig)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer postgres.CloseConnectionPool(pool)
	logger.Info("Database connection established successfully")

	// Initialize message repository
	messageRepo := postgres.NewMessageRepository(pool)

	// Initialize cache service
	logger.Info("Initializing cache service", zap.String("type", "redis"))
	cacheConfig := services.CacheConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		TTL:      cfg.Redis.TTL,
	}
	cacheService := cache.NewRedisService(cacheConfig)
	logger.Info("Cache service initialized successfully")

	// Initialize webhook service
	logger.Info("Initializing webhook service", zap.String("url", cfg.Webhook.URL))
	webhookConfig := services.WebhookConfig{
		URL:              cfg.Webhook.URL,
		AuthToken:        cfg.Webhook.AuthToken,
		Timeout:          cfg.Webhook.Timeout,
		MaxRetries:       cfg.Webhook.MaxRetries,
		RetryBackoffBase: cfg.Webhook.RetryBackoffBase,
	}
	webhookService := webhook.NewWebhookService(webhookConfig)
	logger.Info("Webhook service initialized successfully")

	// Initialize use cases
	logger.Info("Initializing use cases")
	messageMgmtUseCase := usecases.NewMessageManagementService(messageRepo)
	messageProcessingUseCase := usecases.NewMessageProcessingService(
		messageRepo,
		webhookService,
		cacheService,
	)
	logger.Info("Use cases initialized successfully")

	// Initialize scheduler
	logger.Info("Initializing background scheduler",
		zap.Duration("interval", cfg.Scheduler.Interval),
		zap.Int("batch_size", cfg.Scheduler.BatchSize))
	schedulerConfig := schedulerServices.SchedulerConfig{
		Interval:  cfg.Scheduler.Interval,
		BatchSize: cfg.Scheduler.BatchSize,
	}
	scheduler := schedulerServices.NewProcessingScheduler(
		messageProcessingUseCase,
		schedulerConfig,
	)
	logger.Info("Background scheduler initialized successfully")

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	messageHandler := handlers.NewMessageHandler(messageMgmtUseCase, messageProcessingUseCase)
	schedulerHandler := handlers.NewSchedulerHandler(scheduler, cfg.Scheduler.Interval, cfg.Scheduler.BatchSize)

	// Setup routes
	logger.Info("Setting up HTTP routes")
	routerConfig := routes.RouterConfig{
		APIKey: cfg.App.APIKey,
	}
	router := routes.SetupRouter(routerConfig, messageHandler, healthHandler, schedulerHandler)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start scheduler in background
	logger.Info("Starting background scheduler")
	if err := scheduler.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start scheduler", zap.Error(err))
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server",
			zap.String("address", server.Addr),
			zap.String("environment", cfg.App.Environment))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Info("🚀 Message Dispatcher started successfully")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("Shutdown signal received", zap.String("signal", sig.String()))

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop scheduler
	logger.Info("Stopping background scheduler")
	if err := scheduler.Stop(); err != nil {
		logger.Error("Error stopping scheduler", err)
	} else {
		logger.Info("Background scheduler stopped successfully")
	}

	// Shutdown server
	logger.Info("Shutting down HTTP server")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Error shutting down server", err)
	} else {
		logger.Info("HTTP server stopped successfully")
	}

	logger.Info("💤 Message Dispatcher shutdown complete")
}
