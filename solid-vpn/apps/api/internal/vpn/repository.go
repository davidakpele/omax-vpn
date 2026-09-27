package vpn

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

func (r *Repository) AllocateIP(ctx context.Context, serverID uuid.UUID) (netip.Addr, error) {
	var ipStr string
	err := r.db.QueryRow(ctx, `
		UPDATE ip_allocations
		SET allocated = true
		WHERE id = (
			SELECT id FROM ip_allocations
			WHERE server_id = $1 AND allocated = false
			ORDER BY ip_address
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING ip_address
	`, serverID).Scan(&ipStr)
	if errors.Is(err, pgx.ErrNoRows) {
		return netip.Addr{}, ErrNoIPAvailable
	}
	if err != nil {
		return netip.Addr{}, err
	}
	return netip.ParseAddr(ipStr)
}

func (r *Repository) ReleaseIP(ctx context.Context, serverID uuid.UUID, ip netip.Addr) error {
	_, err := r.db.Exec(ctx, `
		UPDATE ip_allocations
		SET allocated = false, peer_id = NULL
		WHERE server_id = $1 AND ip_address = $2
	`, serverID, ip.String())
	return err
}

func (r *Repository) UpsertPeer(ctx context.Context, userID, deviceID, serverID uuid.UUID, publicKey string, assignedIP netip.Addr) (*Peer, error) {
	p := &Peer{}
	var ipStr string
	err := r.db.QueryRow(ctx, `
		INSERT INTO vpn_peers (user_id, device_id, server_id, public_key, assigned_ip)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (server_id, public_key)
		DO UPDATE SET status = 'ACTIVE', updated_at = NOW()
		RETURNING id, user_id, device_id, server_id, public_key, assigned_ip, status, created_at, updated_at
	`, userID, deviceID, serverID, publicKey, assignedIP.String()).Scan(
		&p.ID, &p.UserID, &p.DeviceID, &p.ServerID,
		&p.PublicKey, &ipStr, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	parsed, err := netip.ParseAddr(ipStr)
	if err != nil {
		return nil, err
	}
	p.AssignedIP = parsed
	return p, nil
}

func (r *Repository) GetPeerByDeviceAndServer(ctx context.Context, deviceID, serverID uuid.UUID) (*Peer, error) {
	p := &Peer{}
	var ipStr string
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, device_id, server_id, public_key, assigned_ip, status, created_at, updated_at
		FROM vpn_peers WHERE device_id = $1 AND server_id = $2 AND status = 'ACTIVE'
	`, deviceID, serverID).Scan(
		&p.ID, &p.UserID, &p.DeviceID, &p.ServerID,
		&p.PublicKey, &ipStr, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	parsed, err := netip.ParseAddr(ipStr)
	if err != nil {
		return nil, err
	}
	p.AssignedIP = parsed
	return p, nil
}

func (r *Repository) DisablePeer(ctx context.Context, peerID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE vpn_peers SET status = 'DISABLED', updated_at = NOW() WHERE id = $1
	`, peerID)
	return err
}

func (r *Repository) CreateSession(ctx context.Context, userID, deviceID, serverID uuid.UUID, peerID *uuid.UUID) (*Session, error) {
	s := &Session{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO vpn_sessions (user_id, device_id, server_id, peer_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, device_id, server_id, peer_id, status,
		          bytes_sent, bytes_received, client_ip, started_at, ended_at,
		          created_at, updated_at
	`, userID, deviceID, serverID, peerID).Scan(
		&s.ID, &s.UserID, &s.DeviceID, &s.ServerID, &s.PeerID, &s.Status,
		&s.BytesSent, &s.BytesReceived, &s.ClientIP, &s.StartedAt, &s.EndedAt,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) GetSession(ctx context.Context, sessionID, userID uuid.UUID) (*Session, error) {
	s := &Session{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, device_id, server_id, peer_id, status,
		       bytes_sent, bytes_received, client_ip, started_at, ended_at,
		       created_at, updated_at
		FROM vpn_sessions WHERE id = $1 AND user_id = $2
	`, sessionID, userID).Scan(
		&s.ID, &s.UserID, &s.DeviceID, &s.ServerID, &s.PeerID, &s.Status,
		&s.BytesSent, &s.BytesReceived, &s.ClientIP, &s.StartedAt, &s.EndedAt,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) ListSessions(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]*Session, int, error) {
	q := `
		SELECT id, user_id, device_id, server_id, peer_id, status,
		       bytes_sent, bytes_received, client_ip, started_at, ended_at,
		       created_at, updated_at
		FROM vpn_sessions WHERE user_id = $1
	`
	args := []any{userID}

	if status != "" {
		args = append(args, status)
		q += ` AND status = $2`
	}
	q += ` ORDER BY started_at DESC`

	countQ := `SELECT COUNT(*) FROM vpn_sessions WHERE user_id = $1`
	if status != "" {
		countQ += ` AND status = $2`
	}

	var total int
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	q += ` LIMIT $` + fmt.Sprintf("%d", len(args)-1) + ` OFFSET $` + fmt.Sprintf("%d", len(args))

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		s := &Session{}
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.DeviceID, &s.ServerID, &s.PeerID, &s.Status,
			&s.BytesSent, &s.BytesReceived, &s.ClientIP, &s.StartedAt, &s.EndedAt,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		sessions = append(sessions, s)
	}
	return sessions, total, rows.Err()
}

func (r *Repository) EndSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE vpn_sessions
		SET status = 'ENDED', ended_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'ACTIVE'
	`, sessionID)
	return err
}

func (r *Repository) GetActiveSessionByDevice(ctx context.Context, deviceID uuid.UUID) (*Session, error) {
	s := &Session{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, device_id, server_id, peer_id, status,
		       bytes_sent, bytes_received, client_ip, started_at, ended_at,
		       created_at, updated_at
		FROM vpn_sessions WHERE device_id = $1 AND status = 'ACTIVE'
		LIMIT 1
	`, deviceID).Scan(
		&s.ID, &s.UserID, &s.DeviceID, &s.ServerID, &s.PeerID, &s.Status,
		&s.BytesSent, &s.BytesReceived, &s.ClientIP, &s.StartedAt, &s.EndedAt,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) GetPeerByID(ctx context.Context, id uuid.UUID) (*Peer, error) {
	p := &Peer{}
	var ipStr string
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, device_id, server_id, public_key, assigned_ip, status, created_at, updated_at
		FROM vpn_peers WHERE id = $1
	`, id).Scan(
		&p.ID, &p.UserID, &p.DeviceID, &p.ServerID,
		&p.PublicKey, &ipStr, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	parsed, err := netip.ParseAddr(ipStr)
	if err != nil {
		return nil, err
	}
	p.AssignedIP = parsed
	return p, nil
}
