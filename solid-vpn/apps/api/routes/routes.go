// Package routes wires all HTTP routes together.
package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/solid-vpn/api/internal/health"
	"github.com/solid-vpn/api/internal/middleware"
)

// Options carries the dependencies required to build the router.
type Options struct {
	Logger        *zap.Logger
	HealthHandler *health.Handler
}

// New builds and returns the main HTTP router.
func New(opts Options) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RealIP)
	r.Use(chimw.RequestID)
	r.Use(middleware.RequestLogger(opts.Logger))
	r.Use(middleware.Recoverer(opts.Logger))
	r.Use(chimw.StripSlashes)

	// Health / readiness probes (unauthenticated)
	r.Get("/health", opts.HealthHandler.Live)
	r.Get("/ready", opts.HealthHandler.Ready)

	// Versioned API mount point — handlers will be added in subsequent phases
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"message":"pong"}`))
		})
	})

	return r
}
