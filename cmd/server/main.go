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
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/usecases"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database configuration
	dbConfig := postgres.DefaultDatabaseConfig()

	// Create database connection pool
	pool, err := postgres.NewConnectionPool(ctx, dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer postgres.CloseConnectionPool(pool)

	fmt.Println("✅ Database connected successfully")

	// Create real repository
	messageRepo := postgres.NewMessageRepository(pool)

	// Create webhook service
	webhookConfig := services.WebhookConfig{
		URL:              "https://webhook.site/a25c4f75-0f22-47f4-9def-dbdac00515ae", // Default webhook URL
		Timeout:          30 * time.Second,
		MaxRetries:       3,
		RetryBackoffBase: 100 * time.Millisecond,
	}
	webhookService := webhook.NewWebhookService(webhookConfig)

	// Cache configuration
	cacheConfig := services.CacheConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		TTL:      5 * time.Minute,
	}
	cacheService := cache.NewRedisService(cacheConfig)

	// Create use cases
	messageManagement := usecases.NewMessageManagementService(messageRepo)
	messageProcessing := usecases.NewMessageProcessingService(messageRepo, webhookService, cacheService)

	// Create handlers
	messageHandler := handlers.NewMessageHandler(messageManagement, messageProcessing)
	healthHandler := handlers.NewHealthHandler()

	// Setup router
	routerConfig := routes.RouterConfig{
		APIKey: "test-api-key-123", // In real app, this would come from environment
	}
	router := routes.SetupRouter(routerConfig, messageHandler, healthHandler)

	// Start server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		fmt.Println("🚀 Server starting on :8080")
		fmt.Println("📚 API Documentation:")
		fmt.Println("  GET  /health                     - Health check")
		fmt.Println("  POST /api/v1/messages            - Create message")
		fmt.Println("  GET  /api/v1/messages            - List messages")
		fmt.Println("  GET  /api/v1/messages/{id}       - Get message")
		fmt.Println("  GET  /api/v1/messaging/status    - Processing status")
		fmt.Println("  POST /api/v1/messaging/process   - Manual processing (for testing)")
		fmt.Println("")
		fmt.Println("🔑 Auth: Bearer test-api-key-123")
		fmt.Println("💾 Database: PostgreSQL connected")
		fmt.Println("🗄️  Cache: Redis configured")
		fmt.Println("🔗 Webhook: https://webhook.site/a25c4f75-0f22-47f4-9def-dbdac00515ae")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("🛑 Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	fmt.Println("✅ Server gracefully stopped")
}
