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
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	fmt.Printf("🚀 Starting %s v%s (%s)\n", cfg.App.Name, cfg.App.Version, cfg.App.Environment)

	// Database configuration from config
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

	// Create database connection pool
	pool, err := postgres.NewConnectionPool(ctx, dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer postgres.CloseConnectionPool(pool)

	fmt.Println("✅ Database connected successfully")

	// Create real repository
	messageRepo := postgres.NewMessageRepository(pool)

	// Create webhook service from config
	webhookConfig := services.WebhookConfig{
		URL:              cfg.Webhook.URL,
		AuthToken:        cfg.Webhook.AuthToken,
		Timeout:          cfg.Webhook.Timeout,
		MaxRetries:       cfg.Webhook.MaxRetries,
		RetryBackoffBase: cfg.Webhook.RetryBackoffBase,
	}
	webhookService := webhook.NewWebhookService(webhookConfig)

	// Cache configuration from config
	cacheConfig := services.CacheConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		TTL:      cfg.Redis.TTL,
	}
	cacheService := cache.NewRedisService(cacheConfig)

	// Create use cases
	messageManagement := usecases.NewMessageManagementService(messageRepo)
	messageProcessing := usecases.NewMessageProcessingService(messageRepo, webhookService, cacheService)

	// Create background processing scheduler from config
	schedulerConfig := schedulerServices.SchedulerConfig{
		Interval:  cfg.Scheduler.Interval,
		BatchSize: cfg.Scheduler.BatchSize,
	}
	scheduler := schedulerServices.NewProcessingScheduler(messageProcessing, schedulerConfig)

	// Create handlers
	messageHandler := handlers.NewMessageHandler(messageManagement, messageProcessing)
	healthHandler := handlers.NewHealthHandler()
	schedulerHandler := handlers.NewSchedulerHandler(scheduler, schedulerConfig.Interval, schedulerConfig.BatchSize)

	// Setup router with config
	routerConfig := routes.RouterConfig{
		APIKey: cfg.App.APIKey,
	}
	router := routes.SetupRouter(routerConfig, messageHandler, healthHandler, schedulerHandler)

	// Start background scheduler if enabled
	if cfg.Scheduler.Enabled {
		if err := scheduler.Start(ctx); err != nil {
			log.Fatal("Failed to start background scheduler:", err)
		}
		fmt.Printf("⏰ Background scheduler started (interval: %v, batch size: %d)\n",
			cfg.Scheduler.Interval, cfg.Scheduler.BatchSize)
	} else {
		fmt.Println("⏸️ Background scheduler is disabled")
	}

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("🚀 Server starting on port %d\n", cfg.App.Port)
		fmt.Println("📚 API Documentation:")
		fmt.Println("  GET  /health                     - Health check")
		fmt.Println("  POST /api/v1/messages            - Create message")
		fmt.Println("  GET  /api/v1/messages            - List messages")
		fmt.Println("  GET  /api/v1/messages/{id}       - Get message")
		fmt.Println("  GET  /api/v1/messaging/status    - Processing status")
		fmt.Println("  POST /api/v1/messaging/process   - Manual processing (for testing)")
		fmt.Println("  GET  /api/v1/scheduler/status    - Background scheduler status")
		fmt.Println("  POST /api/v1/scheduler/start     - Start background scheduler")
		fmt.Println("  POST /api/v1/scheduler/stop      - Stop background scheduler")
		fmt.Println("")
		fmt.Printf("🔑 Auth: Bearer %s\n", cfg.App.APIKey)
		fmt.Printf("💾 Database: %s@%s:%d/%s\n", cfg.Database.Username, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database)
		fmt.Printf("🗄️  Cache: %s:%d (DB: %d)\n", cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.DB)
		fmt.Printf("🔗 Webhook: %s\n", cfg.Webhook.URL)
		fmt.Printf("📊 Environment: %s\n", cfg.App.Environment)
		fmt.Printf("📝 Log Level: %s\n", cfg.App.LogLevel)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("🛑 Shutting down server...")

	// Stop background scheduler first
	if cfg.Scheduler.Enabled {
		if err := scheduler.Stop(); err != nil {
			log.Printf("⚠️ Error stopping scheduler: %v", err)
		}
	}

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	fmt.Println("✅ Server gracefully stopped")
}
