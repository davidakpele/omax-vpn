package users

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"

	"github.com/solid-vpn/api/internal/audit"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Service struct {
	repo  *Repository
	audit *audit.Service
}

func NewService(repo *Repository, auditSvc *audit.Service) *Service {
	return &Service{repo: repo, audit: auditSvc}
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*User, error) {
	if req.Email == nil {
		return s.GetByID(ctx, id)
	}

	email := normalizeEmail(*req.Email)
	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}

	taken, err := s.repo.EmailExists(ctx, email, id)
	if err != nil {
		return nil, fmt.Errorf("email check: %w", err)
	}
	if taken {
		return nil, ErrEmailTaken
	}

	u, err := s.repo.UpdateEmail(ctx, id, email)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	s.audit.Log(ctx, audit.ActionUserDeleted, &id, "user", &id, nil)

	return nil
}
