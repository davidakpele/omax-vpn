package vpn

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/solid-vpn/api/internal/audit"
	"github.com/solid-vpn/api/internal/devices"
	"github.com/solid-vpn/api/internal/engine"
	"github.com/solid-vpn/api/internal/servers"
)

type Service struct {
	repo         *Repository
	serverSvc    *servers.Service
	deviceRepo   *devices.Repository
	engineClient *engine.Client
	audit        *audit.Service
	dns          string
}

func NewService(
	repo *Repository,
	serverSvc *servers.Service,
	deviceRepo *devices.Repository,
	engineClient *engine.Client,
	auditSvc *audit.Service,
	dns string,
) *Service {
	return &Service{
		repo:         repo,
		serverSvc:    serverSvc,
		deviceRepo:   deviceRepo,
		engineClient: engineClient,
		audit:        auditSvc,
		dns:          dns,
	}
}

func (s *Service) Connect(ctx context.Context, userID uuid.UUID, req ConnectRequest) (*ConnectResponse, error) {
	device, err := s.deviceRepo.GetByID(ctx, req.DeviceID, userID)
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}
	if device == nil {
		return nil, fmt.Errorf("device not found")
	}

	existing, err := s.repo.GetActiveSessionByDevice(ctx, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("check active session: %w", err)
	}
	if existing != nil {
		return nil, ErrAlreadyActive
	}

	server, err := s.serverSvc.Select(ctx, req.ServerID, "")
	if err != nil {
		return nil, ErrNoServer
	}

	peer, err := s.repo.GetPeerByDeviceAndServer(ctx, req.DeviceID, server.ID)
	if err != nil {
		return nil, fmt.Errorf("get peer: %w", err)
	}

	if peer == nil {
		assignedIP, err := s.repo.AllocateIP(ctx, server.ID)
		if err != nil {
			return nil, err
		}

		peer, err = s.repo.UpsertPeer(ctx, userID, req.DeviceID, server.ID, device.PublicKey, assignedIP)
		if err != nil {
			_ = s.repo.ReleaseIP(ctx, server.ID, assignedIP)
			return nil, fmt.Errorf("create peer: %w", err)
		}
	}

	if _, err := s.engineClient.AddPeer(ctx, engine.AddPeerRequest{
		PeerID:     peer.ID.String(),
		PublicKey:  peer.PublicKey,
		AssignedIP: peer.AssignedIP.String(),
	}); err != nil {
		_ = s.repo.ReleaseIP(ctx, server.ID, peer.AssignedIP)
		_ = s.repo.DisablePeer(ctx, peer.ID)
		return nil, fmt.Errorf("%w: %s", ErrEngineUnavailable, err)
	}

	session, err := s.repo.CreateSession(ctx, userID, req.DeviceID, server.ID, &peer.ID)
	if err != nil {
		_ = s.engineClient.RemovePeer(ctx, peer.PublicKey)
		return nil, fmt.Errorf("create session: %w", err)
	}

	if err := s.serverSvc.IncrementConnections(ctx, server.ID); err != nil {
		return nil, fmt.Errorf("increment connections: %w", err)
	}

	s.audit.Log(ctx, audit.ActionVPNSessionStarted, &userID, "vpn_session", &session.ID, map[string]any{
		"device_id": req.DeviceID.String(),
		"server_id": server.ID.String(),
	})

	cfg := buildWireguardConfig(device.PublicKey, peer.AssignedIP.String(), server, s.dns)

	return &ConnectResponse{
		SessionID:       session.ID,
		WireguardConfig: cfg,
		ServerPublicKey: server.PublicKey,
		AssignedIP:      peer.AssignedIP.String(),
		DNS:             s.dns,
	}, nil
}

func (s *Service) Disconnect(ctx context.Context, userID uuid.UUID, req DisconnectRequest) error {
	session, err := s.repo.GetSession(ctx, req.SessionID, userID)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}
	if session == nil {
		return ErrSessionNotFound
	}

	if err := s.repo.EndSession(ctx, session.ID); err != nil {
		return fmt.Errorf("end session: %w", err)
	}

	if session.PeerID != nil {
		peer, err := s.repo.GetPeerByID(ctx, *session.PeerID)
		if err == nil && peer != nil {
			_ = s.engineClient.RemovePeer(ctx, peer.PublicKey)
		}
		_ = s.repo.DisablePeer(ctx, *session.PeerID)
	}

	if err := s.serverSvc.DecrementConnections(ctx, session.ServerID); err != nil {
		return fmt.Errorf("decrement connections: %w", err)
	}

	s.audit.Log(ctx, audit.ActionVPNSessionEnded, &userID, "vpn_session", &session.ID, map[string]any{
		"server_id": session.ServerID.String(),
	})

	return nil
}

func (s *Service) GetConfig(ctx context.Context, userID, deviceID, serverID uuid.UUID) (*ConfigResponse, error) {
	device, err := s.deviceRepo.GetByID(ctx, deviceID, userID)
	if err != nil || device == nil {
		return nil, fmt.Errorf("device not found")
	}

	server, err := s.serverSvc.GetByID(ctx, serverID)
	if err != nil {
		return nil, err
	}

	peer, err := s.repo.GetPeerByDeviceAndServer(ctx, deviceID, serverID)
	if err != nil || peer == nil {
		return nil, ErrPeerNotFound
	}

	cfg := buildWireguardConfig(device.PublicKey, peer.AssignedIP.String(), server, s.dns)
	return &ConfigResponse{
		WireguardConfig: cfg,
		ServerPublicKey: server.PublicKey,
		AssignedIP:      peer.AssignedIP.String(),
		DNS:             s.dns,
	}, nil
}

func (s *Service) ListSessions(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]*Session, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListSessions(ctx, userID, status, limit, offset)
}

func (s *Service) GetSession(ctx context.Context, sessionID, userID uuid.UUID) (*Session, error) {
	session, err := s.repo.GetSession(ctx, sessionID, userID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func buildWireguardConfig(clientPublicKey, assignedIP string, server *servers.Server, dns string) string {
	return fmt.Sprintf(`[Interface]
Address = %s/32
DNS = %s

[Peer]
PublicKey = %s
Endpoint = %s:%d
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
`, assignedIP, dns, server.PublicKey, server.PublicIP.String(), server.WireguardPort)
}
