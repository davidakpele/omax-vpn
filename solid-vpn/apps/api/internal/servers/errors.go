package servers

import "errors"

var (
	ErrNotFound     = errors.New("server not found")
	ErrNameTaken    = errors.New("server name already exists")
	ErrNoHealthy    = errors.New("no healthy VPN server available")
)
