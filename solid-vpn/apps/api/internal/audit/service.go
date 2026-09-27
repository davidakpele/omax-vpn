package audit

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service struct {
	repo *Repository
	log  *zap.Logger
}

func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) Log(ctx context.Context, action Action, actorID *uuid.UUID, resourceType string, resourceID *uuid.UUID, metadata map[string]any) {
	e := Entry{
		ActorID:      actorID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Metadata:     metadata,
	}
	if err := s.repo.Write(ctx, e); err != nil {
		s.log.Error("audit write failed",
			zap.String("action", string(action)),
			zap.Error(err),
		)
	}
}
