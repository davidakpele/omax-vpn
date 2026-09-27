package vpn

import "errors"

var (
	ErrNoServer        = errors.New("no healthy VPN server available")
	ErrPeerNotFound    = errors.New("VPN peer not found")
	ErrSessionNotFound = errors.New("VPN session not found")
	ErrAlreadyActive   = errors.New("device already has an active session")
	ErrNoIPAvailable   = errors.New("no IP address available in pool")
	ErrEngineUnavailable = errors.New("VPN engine is unavailable")
)
