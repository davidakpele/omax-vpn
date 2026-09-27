package servers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, country string, regionID *uuid.UUID) ([]*Server, error) {
	list, err := s.repo.List(ctx, country, regionID)
	if err != nil {
		return nil, fmt.Errorf("list servers: %w", err)
	}
	if list == nil {
		list = []*Server{}
	}
	return list, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Server, error) {
	srv, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get server: %w", err)
	}
	if srv == nil {
		return nil, ErrNotFound
	}
	return srv, nil
}

func (s *Service) ListRegions(ctx context.Context) ([]*Region, error) {
	regions, err := s.repo.ListRegions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list regions: %w", err)
	}
	if regions == nil {
		regions = []*Region{}
	}
	return regions, nil
}

func (s *Service) Select(ctx context.Context, preferredID *uuid.UUID, country string) (*Server, error) {
	if preferredID != nil {
		srv, err := s.repo.GetByID(ctx, *preferredID)
		if err != nil {
			return nil, fmt.Errorf("get preferred server: %w", err)
		}
		if srv != nil && srv.IsSelectable() && srv.ActiveConnections < srv.Capacity {
			return srv, nil
		}
	}

	candidates, err := s.repo.ListSelectable(ctx)
	if err != nil {
		return nil, fmt.Errorf("list selectable: %w", err)
	}

	if country != "" {
		var filtered []*Server
		for _, c := range candidates {
			if c.Country == country {
				filtered = append(filtered, c)
			}
		}
		if len(filtered) > 0 {
			candidates = filtered
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoHealthy
	}

	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.Load() < best.Load() {
			best = c
		}
	}
	return best, nil
}

func (s *Service) IncrementConnections(ctx context.Context, id uuid.UUID) error {
	return s.repo.IncrementConnections(ctx, id)
}

func (s *Service) DecrementConnections(ctx context.Context, id uuid.UUID) error {
	return s.repo.DecrementConnections(ctx, id)
}
