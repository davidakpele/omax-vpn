package auth

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, "INVALID_BODY", "Request body is not valid JSON.")
		return
	}

	user, err := h.svc.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailTaken):
			respondErr(w, http.StatusConflict, "EMAIL_TAKEN", err.Error())
		case errors.Is(err, ErrWeakPassword):
			respondErr(w, http.StatusBadRequest, "WEAK_PASSWORD", err.Error())
		case errors.Is(err, ErrInvalidEmail):
			respondErr(w, http.StatusBadRequest, "INVALID_EMAIL", err.Error())
		default:
			respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		}
		return
	}

	respond(w, http.StatusCreated, map[string]any{
		"user_id":    user.ID,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, "INVALID_BODY", "Request body is not valid JSON.")
		return
	}

	tokens, err := h.svc.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCreds):
			respondErr(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error())
		case errors.Is(err, ErrAccountSuspended):
			respondErr(w, http.StatusForbidden, "ACCOUNT_SUSPENDED", err.Error())
		default:
			respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		}
		return
	}

	respond(w, http.StatusOK, tokens)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, "INVALID_BODY", "Request body is not valid JSON.")
		return
	}

	resp, err := h.svc.Refresh(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			respondErr(w, http.StatusUnauthorized, "INVALID_TOKEN", err.Error())
			return
		}
		respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	respond(w, http.StatusOK, resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_ = h.svc.Logout(r.Context(), req.RefreshToken)
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
