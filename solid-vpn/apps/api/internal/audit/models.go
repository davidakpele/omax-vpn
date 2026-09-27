package audit

import (
	"time"

	"github.com/google/uuid"
)

type Action string

const (
	ActionUserCreated       Action = "USER_CREATED"
	ActionUserLogin         Action = "USER_LOGIN"
	ActionUserLoginFailed   Action = "USER_LOGIN_FAILED"
	ActionUserDeleted       Action = "USER_DELETED"
	ActionDeviceRegistered  Action = "DEVICE_REGISTERED"
	ActionDeviceDeleted     Action = "DEVICE_DELETED"
	ActionVPNSessionStarted Action = "VPN_SESSION_STARTED"
	ActionVPNSessionEnded   Action = "VPN_SESSION_ENDED"
)

type Entry struct {
	ID           uuid.UUID
	ActorID      *uuid.UUID
	Action       Action
	ResourceType string
	ResourceID   *uuid.UUID
	Metadata     map[string]any
	CreatedAt    time.Time
}
