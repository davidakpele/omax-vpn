package devices

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
	userID := userIDFromCtx(r)
	if userID == uuid.Nil {
		respondErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	list, err := h.svc.List(r.Context(), userID)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	respond(w, http.StatusOK, map[string]any{"devices": list})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)
	if userID == uuid.Nil {
		respondErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, "INVALID_BODY", "Request body is not valid JSON.")
		return
	}

	d, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrKeyTaken):
			respondErr(w, http.StatusConflict, "KEY_TAKEN", err.Error())
		case errors.Is(err, ErrInvalidKey):
			respondErr(w, http.StatusBadRequest, "INVALID_KEY", err.Error())
		case errors.Is(err, ErrInvalidName):
			respondErr(w, http.StatusBadRequest, "INVALID_NAME", err.Error())
		default:
			respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		}
		return
	}

	respond(w, http.StatusCreated, d)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)
	if userID == uuid.Nil {
		respondErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	deviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "INVALID_ID", "Device ID must be a valid UUID.")
		return
	}

	d, err := h.svc.GetByID(r.Context(), deviceID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respondErr(w, http.StatusNotFound, "DEVICE_NOT_FOUND", err.Error())
			return
		}
		respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	respond(w, http.StatusOK, d)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)
	if userID == uuid.Nil {
		respondErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	deviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "INVALID_ID", "Device ID must be a valid UUID.")
		return
	}

	if err := h.svc.Delete(r.Context(), deviceID, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			respondErr(w, http.StatusNotFound, "DEVICE_NOT_FOUND", err.Error())
			return
		}
		respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
