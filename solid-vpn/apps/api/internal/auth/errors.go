package auth

import "errors"

var (
	ErrEmailTaken       = errors.New("email already registered")
	ErrInvalidCreds     = errors.New("invalid email or password")
	ErrInvalidToken     = errors.New("invalid or expired token")
	ErrAccountSuspended = errors.New("account is suspended")
	ErrWeakPassword     = errors.New("password must be at least 12 characters")
	ErrInvalidEmail     = errors.New("invalid email address")
)
