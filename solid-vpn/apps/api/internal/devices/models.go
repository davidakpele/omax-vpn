package devices

import (
	"time"

	"github.com/google/uuid"
)

type DeviceStatus string

const (
	StatusActive  DeviceStatus = "ACTIVE"
	StatusRevoked DeviceStatus = "REVOKED"
)

type Device struct {
	ID        uuid.UUID    `json:"id"`
	UserID    uuid.UUID    `json:"user_id"`
	Name      string       `json:"name"`
	PublicKey string       `json:"public_key"`
	Status    DeviceStatus `json:"status"`
	LastSeen  *time.Time   `json:"last_seen"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type CreateRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}
