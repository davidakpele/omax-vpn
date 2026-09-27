package devices

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/solid-vpn/api/internal/audit"
)

type Service struct {
	repo  *Repository
	audit *audit.Service
}

func NewService(repo *Repository, auditSvc *audit.Service) *Service {
	return &Service{repo: repo, audit: auditSvc}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateRequest) (*Device, error) {
	name := strings.TrimSpace(req.Name)
	if len(name) == 0 || len(name) > 100 {
		return nil, ErrInvalidName
	}
	if !isValidPublicKey(req.PublicKey) {
		return nil, ErrInvalidKey
	}

	exists, err := s.repo.PublicKeyExists(ctx, req.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("key check: %w", err)
	}
	if exists {
		return nil, ErrKeyTaken
	}

	d, err := s.repo.Create(ctx, userID, name, req.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("create device: %w", err)
	}

	s.audit.Log(ctx, audit.ActionDeviceRegistered, &userID, "device", &d.ID, map[string]any{
		"name": d.Name,
	})

	return d, nil
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]*Device, error) {
	list, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	if list == nil {
		list = []*Device{}
	}
	return list, nil
}

func (s *Service) GetByID(ctx context.Context, id, userID uuid.UUID) (*Device, error) {
	d, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}
	return d, nil
}

func (s *Service) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if err := s.repo.Revoke(ctx, id, userID); err != nil {
		return err
	}

	s.audit.Log(ctx, audit.ActionDeviceDeleted, &userID, "device", &id, nil)

	return nil
}

func isValidPublicKey(key string) bool {
	if len(key) != 44 {
		return false
	}
	for _, c := range key {
		if !isBase64Char(c) {
			return false
		}
	}
	return true
}

func isBase64Char(c rune) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '+' || c == '/' || c == '='
}
