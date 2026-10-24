package httptransport

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/endorsain/nictofilia-go-api/internal/config"
	httpmiddleware "github.com/endorsain/nictofilia-go-api/internal/transport/http/middleware"
)

func NewRouter(cfg config.ServerConfig) *chi.Mux {
	r := chi.NewRouter()
	UseCommonMiddleware(r, cfg)
	return r
}

func UseCommonMiddleware(r *chi.Mux, cfg config.ServerConfig) {
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(corsMiddleware(cfg.CORS))
}

// corsMiddleware never combines a wildcard origin with credentials, since
// browsers reject that combination.
func corsMiddleware(cfg config.CORSConfig) func(http.Handler) http.Handler {
	allowCredentials := cfg.AllowCredentials
	for _, origin := range cfg.AllowedOrigins {
		if origin == "*" {
			allowCredentials = false
			break
		}
	}

	return cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		AllowCredentials: allowCredentials,
	})
}

// Handle registers an error-returning handler, routing any error it returns
// through the centralized HTTP error middleware (see transport/http/middleware).
func Handle(fn httpmiddleware.HandlerFunc) http.HandlerFunc {
	return httpmiddleware.Errors(fn)
}

// SetErrorLogger configures the logger used to record request failures
// handled by Handle. It should be called once, from the composition root,
// before the router starts serving requests.
func SetErrorLogger(logger *slog.Logger) {
	httpmiddleware.SetLogger(logger)
}
