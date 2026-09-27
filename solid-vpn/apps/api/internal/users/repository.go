package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, email, role, status, created_at, updated_at
		FROM users WHERE id = $1 AND status != 'DELETED'
	`, id).Scan(&u.ID, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) EmailExists(ctx context.Context, email string, excludeID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM users WHERE email = $1 AND status != 'DELETED' AND id != $2
		)
	`, email, excludeID).Scan(&exists)
	return exists, err
}

func (r *Repository) UpdateEmail(ctx context.Context, id uuid.UUID, email string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx, `
		UPDATE users SET email = $1, updated_at = NOW()
		WHERE id = $2 AND status != 'DELETED'
		RETURNING id, email, role, status, created_at, updated_at
	`, email, id).Scan(&u.ID, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE users SET status = 'DELETED', updated_at = NOW() WHERE id = $1 AND status != 'DELETED'
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no rows updated")
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
