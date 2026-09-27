package users

import (
	"time"

	"github.com/google/uuid"
)

type UserRole   string
type UserStatus string

const (
	RoleUser   UserRole = "USER"
	RoleAdmin  UserRole = "ADMIN"
	RoleSystem UserRole = "SYSTEM"

	StatusActive    UserStatus = "ACTIVE"
	StatusSuspended UserStatus = "SUSPENDED"
	StatusDeleted   UserStatus = "DELETED"
)

type User struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Role      UserRole   `json:"role"`
	Status    UserStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type UpdateRequest struct {
	Email *string `json:"email"`
}
