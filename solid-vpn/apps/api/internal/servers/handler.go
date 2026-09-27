package servers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	var regionID *uuid.UUID
	if raw := r.URL.Query().Get("region_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			respondErr(w, http.StatusBadRequest, "INVALID_REGION_ID", "region_id must be a valid UUID.")
			return
		}
		regionID = &id
	}

	list, err := h.svc.List(r.Context(), country, regionID)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	respond(w, http.StatusOK, map[string]any{"servers": list})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "INVALID_ID", "Server ID must be a valid UUID.")
		return
	}

	srv, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respondErr(w, http.StatusNotFound, "SERVER_NOT_FOUND", err.Error())
			return
		}
		respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	respond(w, http.StatusOK, srv)
}

func (h *Handler) ListRegions(w http.ResponseWriter, r *http.Request) {
	regions, err := h.svc.ListRegions(r.Context())
	if err != nil {
		respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}
	respond(w, http.StatusOK, map[string]any{"regions": regions})
}

func respond(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func respondErr(w http.ResponseWriter, code int, errCode, msg string) {
	respond(w, code, map[string]any{
		"error": map[string]string{
			"code":    errCode,
			"message": msg,
		},
	})
}
