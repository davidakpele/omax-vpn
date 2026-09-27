package servers

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

type ServerStatus string

const (
	StatusHealthy     ServerStatus = "HEALTHY"
	StatusDegraded    ServerStatus = "DEGRADED"
	StatusMaintenance ServerStatus = "MAINTENANCE"
	StatusOffline     ServerStatus = "OFFLINE"
)

type Region struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Server struct {
	ID                uuid.UUID    `json:"id"`
	RegionID          uuid.UUID    `json:"region_id"`
	Name              string       `json:"name"`
	Hostname          string       `json:"hostname"`
	PublicIP          netip.Addr   `json:"public_ip"`
	Country           string       `json:"country"`
	City              string       `json:"city"`
	Provider          string       `json:"provider"`
	Status            ServerStatus `json:"status"`
	Capacity          int          `json:"capacity"`
	ActiveConnections int          `json:"active_connections"`
	WireguardPort     int          `json:"wireguard_port"`
	PublicKey         string       `json:"public_key"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}

func (s *Server) Load() float64 {
	if s.Capacity == 0 {
		return 1.0
	}
	return float64(s.ActiveConnections) / float64(s.Capacity)
}

func (s *Server) IsSelectable() bool {
	return s.Status == StatusHealthy || s.Status == StatusDegraded
}
