package devices

import "errors"

var (
	ErrNotFound      = errors.New("device not found")
	ErrKeyTaken      = errors.New("public key already registered")
	ErrInvalidKey    = errors.New("invalid WireGuard public key")
	ErrInvalidName   = errors.New("device name must be between 1 and 100 characters")
)
