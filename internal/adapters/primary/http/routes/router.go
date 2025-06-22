package routes

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/svbnbyrk/go-message-dispatcher/docs" // Import generated docs
	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/primary/http/handlers"
	httpMiddleware "github.com/svbnbyrk/go-message-dispatcher/internal/adapters/primary/http/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// RouterConfig contains configuration for the router
type RouterConfig struct {
	APIKey string
}

// SetupRouter creates and configures the Chi router
func SetupRouter(
	config RouterConfig,
	messageHandler *handlers.MessageHandler,
	healthHandler *handlers.HealthHandler,
	schedulerHandler *handlers.SchedulerHandler,
) *chi.Mux {
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Global rate limiting (60 requests per minute per IP)
	r.Use(httpMiddleware.RateLimitMiddleware(httpMiddleware.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 60,
		CleanupInterval:   5 * time.Minute,
	}))

	r.Use(httpMiddleware.RequestLogging()) // Structured logging with correlation IDs
	r.Use(httpMiddleware.PanicRecovery())  // Panic recovery with structured logging
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware (simple)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Correlation-ID")

			if r.Method == "OPTIONS" {
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Health endpoint (no auth required)
	r.Get("/health", healthHandler.Health)

	// Swagger documentation (development only)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	// API routes with authentication
	r.Route("/api/v1", func(r chi.Router) {
		// Apply auth middleware
		r.Use(httpMiddleware.ClientCredentialsAuth(httpMiddleware.AuthConfig{
			APIKey: config.APIKey,
		}))

		// Message routes
		r.Route("/messages", func(r chi.Router) {
			r.Post("/", messageHandler.CreateMessage)
			r.Get("/", messageHandler.ListMessages)
			r.Get("/{id}", messageHandler.GetMessage)
		})

		// Messaging status routes
		r.Route("/messaging", func(r chi.Router) {
			r.Get("/status", messageHandler.GetProcessingStatus)
			r.Post("/process", messageHandler.ProcessMessages)
		})

		// Scheduler management routes
		r.Route("/scheduler", func(r chi.Router) {
			r.Get("/status", schedulerHandler.GetSchedulerStatus)
			r.Post("/start", schedulerHandler.StartScheduler)
			r.Post("/stop", schedulerHandler.StopScheduler)
		})
	})

	return r
}
