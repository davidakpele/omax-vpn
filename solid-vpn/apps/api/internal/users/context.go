package users

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/solid-vpn/api/internal/middleware"
)

func userIDFromCtx(r *http.Request) uuid.UUID {
	v := r.Context().Value(middleware.ContextKeyUserID)
	if v == nil {
		return uuid.Nil
	}
	id, _ := v.(uuid.UUID)
	return id
}
