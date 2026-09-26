// Package health provides HTTP handlers for liveness and readiness probes.
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// DB is the minimal interface the health handler needs to check the database.
type DB interface {
	Ping(ctx context.Context) error
}

// Handler holds dependencies for health endpoints.
type Handler struct {
	db DB
}

// NewHandler creates a health Handler.
// db may be nil during startup; the readiness check will report not-ready.
func NewHandler(db DB) *Handler {
	return &Handler{db: db}
}

type healthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

type readyResponse struct {
	Status   string            `json:"status"`
	Checks   map[string]string `json:"checks"`
	Timestamp time.Time        `json:"timestamp"`
}

// Live handles GET /health — always returns 200 while the process is running.
func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, healthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Service:   "vpn-api",
	})
}

// Ready handles GET /ready — returns 200 only when all dependencies are up.
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)
	allOK := true

	// Database check
	if h.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := h.db.Ping(ctx); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
			allOK = false
		} else {
			checks["database"] = "healthy"
		}
	} else {
		checks["database"] = "not configured"
		allOK = false
	}

	status := http.StatusOK
	statusText := "ready"
	if !allOK {
		status = http.StatusServiceUnavailable
		statusText = "not ready"
	}

	respond(w, status, readyResponse{
		Status:    statusText,
		Checks:    checks,
		Timestamp: time.Now().UTC(),
	})
}

func respond(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
