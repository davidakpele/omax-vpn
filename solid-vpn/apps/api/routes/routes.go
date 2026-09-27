package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/solid-vpn/api/internal/auth"
	"github.com/solid-vpn/api/internal/devices"
	"github.com/solid-vpn/api/internal/health"
	"github.com/solid-vpn/api/internal/middleware"
	"github.com/solid-vpn/api/internal/servers"
	"github.com/solid-vpn/api/internal/users"
	"github.com/solid-vpn/api/internal/vpn"
)

type Options struct {
	Logger        *zap.Logger
	HealthHandler *health.Handler
	AuthHandler   *auth.Handler
	AuthService   *auth.Service
	UserHandler   *users.Handler
	DeviceHandler *devices.Handler
	ServerHandler *servers.Handler
	VPNHandler    *vpn.Handler
}

func New(opts Options) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RealIP)
	r.Use(chimw.RequestID)
	r.Use(middleware.RequestLogger(opts.Logger))
	r.Use(middleware.Recoverer(opts.Logger))
	r.Use(chimw.StripSlashes)

	r.Get("/health", opts.HealthHandler.Live)
	r.Get("/ready", opts.HealthHandler.Ready)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", opts.AuthHandler.Register)
			r.Post("/login", opts.AuthHandler.Login)
			r.Post("/refresh", opts.AuthHandler.Refresh)
			r.Post("/logout", opts.AuthHandler.Logout)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(opts.AuthService, opts.Logger))

			r.Route("/users", func(r chi.Router) {
				r.Get("/me", opts.UserHandler.GetMe)
				r.Patch("/me", opts.UserHandler.UpdateMe)
				r.Delete("/me", opts.UserHandler.DeleteMe)
			})

			r.Route("/devices", func(r chi.Router) {
				r.Get("/", opts.DeviceHandler.List)
				r.Post("/", opts.DeviceHandler.Create)
				r.Get("/{id}", opts.DeviceHandler.Get)
				r.Delete("/{id}", opts.DeviceHandler.Delete)
			})

			r.Route("/vpn", func(r chi.Router) {
				r.Get("/servers", opts.ServerHandler.List)
				r.Get("/servers/{id}", opts.ServerHandler.Get)
				r.Get("/regions", opts.ServerHandler.ListRegions)

				r.Post("/connect", opts.VPNHandler.Connect)
				r.Post("/disconnect", opts.VPNHandler.Disconnect)
				r.Get("/config", opts.VPNHandler.GetConfig)

				r.Get("/sessions", opts.VPNHandler.ListSessions)
				r.Get("/sessions/{id}", opts.VPNHandler.GetSession)
			})
		})
	})

	return r
}
