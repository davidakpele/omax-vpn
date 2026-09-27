package servers

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

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

func (r *Repository) ListRegions(ctx context.Context) ([]*Region, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, code, name, created_at FROM vpn_regions ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []*Region
	for rows.Next() {
		rg := &Region{}
		if err := rows.Scan(&rg.ID, &rg.Code, &rg.Name, &rg.CreatedAt); err != nil {
			return nil, err
		}
		regions = append(regions, rg)
	}
	return regions, rows.Err()
}

func (r *Repository) List(ctx context.Context, country string, regionID *uuid.UUID) ([]*Server, error) {
	q := `
		SELECT id, region_id, name, hostname, public_ip, country, city, provider,
		       status, capacity, active_connections, wireguard_port, public_key,
		       created_at, updated_at
		FROM vpn_servers
		WHERE status != 'OFFLINE'
	`
	args := []any{}

	if country != "" {
		args = append(args, country)
		q += ` AND country = $` + itoa(len(args))
	}
	if regionID != nil {
		args = append(args, *regionID)
		q += ` AND region_id = $` + itoa(len(args))
	}

	q += ` ORDER BY active_connections ASC`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanServers(rows)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Server, error) {
	s := &Server{}
	err := r.db.QueryRow(ctx, `
		SELECT id, region_id, name, hostname, public_ip, country, city, provider,
		       status, capacity, active_connections, wireguard_port, public_key,
		       created_at, updated_at
		FROM vpn_servers WHERE id = $1
	`, id).Scan(
		&s.ID, &s.RegionID, &s.Name, &s.Hostname, &s.PublicIP, &s.Country, &s.City,
		&s.Provider, &s.Status, &s.Capacity, &s.ActiveConnections,
		&s.WireguardPort, &s.PublicKey, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return s, err
}

func (r *Repository) ListSelectable(ctx context.Context) ([]*Server, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, region_id, name, hostname, public_ip, country, city, provider,
		       status, capacity, active_connections, wireguard_port, public_key,
		       created_at, updated_at
		FROM vpn_servers
		WHERE status IN ('HEALTHY', 'DEGRADED')
		  AND active_connections < capacity
		ORDER BY active_connections ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanServers(rows)
}

func (r *Repository) IncrementConnections(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE vpn_servers
		SET active_connections = active_connections + 1, updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

func (r *Repository) DecrementConnections(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE vpn_servers
		SET active_connections = GREATEST(active_connections - 1, 0), updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

func scanServers(rows pgx.Rows) ([]*Server, error) {
	var servers []*Server
	for rows.Next() {
		s := &Server{}
		if err := rows.Scan(
			&s.ID, &s.RegionID, &s.Name, &s.Hostname, &s.PublicIP, &s.Country, &s.City,
			&s.Provider, &s.Status, &s.Capacity, &s.ActiveConnections,
			&s.WireguardPort, &s.PublicKey, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}
	return servers, rows.Err()
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

var _ = netip.Addr{}
