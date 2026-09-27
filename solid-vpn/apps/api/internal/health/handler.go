package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type DB interface {
	Ping(ctx context.Context) error
}

type Engine interface {
	Health(ctx context.Context) error
}

type Handler struct {
	db     DB
	engine Engine
}

func NewHandler(db DB, eng Engine) *Handler {
	return &Handler{db: db, engine: eng}
}

type healthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

type readyResponse struct {
	Status    string            `json:"status"`
	Checks    map[string]string `json:"checks"`
	Timestamp time.Time         `json:"timestamp"`
}

func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, healthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Service:   "vpn-api",
	})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)
	allOK := true

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

	if h.engine != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := h.engine.Health(ctx); err != nil {
			checks["vpn_engine"] = "unhealthy: " + err.Error()
			allOK = false
		} else {
			checks["vpn_engine"] = "healthy"
		}
	} else {
		checks["vpn_engine"] = "not configured"
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
