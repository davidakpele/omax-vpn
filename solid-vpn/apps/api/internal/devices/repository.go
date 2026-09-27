package devices

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, name, publicKey string) (*Device, error) {
	d := &Device{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO devices (user_id, name, public_key)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, name, public_key, status, last_seen, created_at, updated_at
	`, userID, name, publicKey).Scan(
		&d.ID, &d.UserID, &d.Name, &d.PublicKey,
		&d.Status, &d.LastSeen, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*Device, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, name, public_key, status, last_seen, created_at, updated_at
		FROM devices WHERE user_id = $1 AND status = 'ACTIVE'
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*Device
	for rows.Next() {
		d := &Device{}
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.Name, &d.PublicKey,
			&d.Status, &d.LastSeen, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id, userID uuid.UUID) (*Device, error) {
	d := &Device{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, name, public_key, status, last_seen, created_at, updated_at
		FROM devices WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(
		&d.ID, &d.UserID, &d.Name, &d.PublicKey,
		&d.Status, &d.LastSeen, &d.CreatedAt, &d.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *Repository) PublicKeyExists(ctx context.Context, publicKey string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM devices WHERE public_key = $1 AND status = 'ACTIVE')
	`, publicKey).Scan(&exists)
	return exists, err
}

func (r *Repository) Revoke(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE devices SET status = 'REVOKED', updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND status = 'ACTIVE'
	`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
