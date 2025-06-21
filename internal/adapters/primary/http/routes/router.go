package routes

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/primary/http/handlers"
	httpMiddleware "github.com/svbnbyrk/go-message-dispatcher/internal/adapters/primary/http/middleware"
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
) *chi.Mux {
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware (simple)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

			if r.Method == "OPTIONS" {
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Health check (no auth required)
	r.Get("/health", healthHandler.Health)

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
	})

	return r
}
