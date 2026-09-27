package vpn

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/solid-vpn/api/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)
	if userID == uuid.Nil {
		respondErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	var req ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, "INVALID_BODY", "Request body is not valid JSON.")
		return
	}

	resp, err := h.svc.Connect(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNoServer):
			respondErr(w, http.StatusServiceUnavailable, "NO_SERVER_AVAILABLE", err.Error())
		case errors.Is(err, ErrAlreadyActive):
			respondErr(w, http.StatusConflict, "SESSION_ALREADY_ACTIVE", err.Error())
		case errors.Is(err, ErrNoIPAvailable):
			respondErr(w, http.StatusServiceUnavailable, "NO_IP_AVAILABLE", err.Error())
		default:
			respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		}
		return
	}

	respond(w, http.StatusOK, resp)
}

func (h *Handler) Disconnect(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)
	if userID == uuid.Nil {
		respondErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	var req DisconnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, "INVALID_BODY", "Request body is not valid JSON.")
		return
	}

	if err := h.svc.Disconnect(r.Context(), userID, req); err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			respondErr(w, http.StatusNotFound, "SESSION_NOT_FOUND", err.Error())
			return
		}
		respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)
	if userID == uuid.Nil {
		respondErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	deviceID, err := uuid.Parse(r.URL.Query().Get("device_id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "INVALID_DEVICE_ID", "device_id must be a valid UUID.")
		return
	}

	serverID, err := uuid.Parse(r.URL.Query().Get("server_id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "INVALID_SERVER_ID", "server_id must be a valid UUID.")
		return
	}

	cfg, err := h.svc.GetConfig(r.Context(), userID, deviceID, serverID)
	if err != nil {
		if errors.Is(err, ErrPeerNotFound) {
			respondErr(w, http.StatusNotFound, "PEER_NOT_FOUND", err.Error())
			return
		}
		respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	respond(w, http.StatusOK, cfg)
}

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)
	if userID == uuid.Nil {
		respondErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	sessions, total, err := h.svc.ListSessions(r.Context(), userID, status, limit, offset)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	respond(w, http.StatusOK, map[string]any{
		"sessions": sessions,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)
	if userID == uuid.Nil {
		respondErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "INVALID_ID", "Session ID must be a valid UUID.")
		return
	}

	session, err := h.svc.GetSession(r.Context(), sessionID, userID)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			respondErr(w, http.StatusNotFound, "SESSION_NOT_FOUND", err.Error())
			return
		}
		respondErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	respond(w, http.StatusOK, session)
}

func userIDFromCtx(r *http.Request) uuid.UUID {
	v := r.Context().Value(middleware.ContextKeyUserID)
	if v == nil {
		return uuid.Nil
	}
	id, _ := v.(uuid.UUID)
	return id
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
