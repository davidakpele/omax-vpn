package vpn

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

type PeerStatus string
type SessionStatus string

const (
	PeerActive   PeerStatus = "ACTIVE"
	PeerDisabled PeerStatus = "DISABLED"
	PeerRemoved  PeerStatus = "REMOVED"

	SessionActive  SessionStatus = "ACTIVE"
	SessionEnded   SessionStatus = "ENDED"
	SessionExpired SessionStatus = "EXPIRED"
	SessionError   SessionStatus = "ERROR"
)

type Peer struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	DeviceID   uuid.UUID  `json:"device_id"`
	ServerID   uuid.UUID  `json:"server_id"`
	PublicKey  string     `json:"public_key"`
	AssignedIP netip.Addr `json:"assigned_ip"`
	Status     PeerStatus `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type Session struct {
	ID            uuid.UUID     `json:"id"`
	UserID        uuid.UUID     `json:"user_id"`
	DeviceID      uuid.UUID     `json:"device_id"`
	ServerID      uuid.UUID     `json:"server_id"`
	PeerID        *uuid.UUID    `json:"peer_id"`
	Status        SessionStatus `json:"status"`
	BytesSent     int64         `json:"bytes_sent"`
	BytesReceived int64         `json:"bytes_received"`
	ClientIP      *netip.Addr   `json:"client_ip,omitempty"`
	StartedAt     time.Time     `json:"started_at"`
	EndedAt       *time.Time    `json:"ended_at"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type ConnectRequest struct {
	DeviceID uuid.UUID  `json:"device_id"`
	ServerID *uuid.UUID `json:"server_id"`
}

type ConnectResponse struct {
	SessionID      uuid.UUID `json:"session_id"`
	WireguardConfig string   `json:"wireguard_config"`
	ServerPublicKey string   `json:"server_public_key"`
	AssignedIP      string   `json:"assigned_ip"`
	DNS             string   `json:"dns"`
}

type DisconnectRequest struct {
	SessionID uuid.UUID `json:"session_id"`
}

type ConfigResponse struct {
	WireguardConfig string `json:"wireguard_config"`
	ServerPublicKey string `json:"server_public_key"`
	AssignedIP      string `json:"assigned_ip"`
	DNS             string `json:"dns"`
}
