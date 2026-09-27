package users

import "errors"

var (
	ErrNotFound     = errors.New("user not found")
	ErrEmailTaken   = errors.New("email already registered")
	ErrInvalidEmail = errors.New("invalid email address")
)
