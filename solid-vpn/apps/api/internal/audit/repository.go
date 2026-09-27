package audit

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Write(ctx context.Context, e Entry) error {
	meta, err := json.Marshal(e.Metadata)
	if err != nil {
		meta = []byte("{}")
	}

	var resourceID *uuid.UUID
	if e.ResourceID != nil {
		id := *e.ResourceID
		resourceID = &id
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO audit_logs (actor_id, action, resource_type, resource_id, metadata)
		VALUES ($1, $2, $3, $4, $5)
	`, e.ActorID, string(e.Action), e.ResourceType, resourceID, meta)
	return err
}
