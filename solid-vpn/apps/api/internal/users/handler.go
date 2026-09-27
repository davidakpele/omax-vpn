package users

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)
	if userID == uuid.Nil {
		respondErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	u, err := h.svc.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respondErr(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
			return
		}
		respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	respond(w, http.StatusOK, u)
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)
	if userID == uuid.Nil {
		respondErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, "INVALID_BODY", "Request body is not valid JSON.")
		return
	}

	u, err := h.svc.Update(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailTaken):
			respondErr(w, http.StatusConflict, "EMAIL_TAKEN", err.Error())
		case errors.Is(err, ErrInvalidEmail):
			respondErr(w, http.StatusBadRequest, "INVALID_EMAIL", err.Error())
		case errors.Is(err, ErrNotFound):
			respondErr(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
		default:
			respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		}
		return
	}

	respond(w, http.StatusOK, u)
}

func (h *Handler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)
	if userID == uuid.Nil {
		respondErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	if err := h.svc.Delete(r.Context(), userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			respondErr(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
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
