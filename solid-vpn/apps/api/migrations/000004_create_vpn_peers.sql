CREATE TYPE peer_status AS ENUM ('ACTIVE', 'DISABLED', 'REMOVED');

CREATE TABLE ip_allocations (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id  UUID        NOT NULL REFERENCES vpn_servers (id) ON DELETE CASCADE,
    ip_address INET        NOT NULL,
    peer_id    UUID,
    allocated  BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (server_id, ip_address)
);

CREATE INDEX idx_ip_allocations_server_id ON ip_allocations (server_id);
CREATE INDEX idx_ip_allocations_allocated ON ip_allocations (allocated);

CREATE TABLE vpn_peers (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users (id)       ON DELETE CASCADE,
    device_id   UUID        NOT NULL REFERENCES devices (id)     ON DELETE CASCADE,
    server_id   UUID        NOT NULL REFERENCES vpn_servers (id) ON DELETE CASCADE,
    public_key  TEXT        NOT NULL,
    assigned_ip INET        NOT NULL,
    status      peer_status NOT NULL DEFAULT 'ACTIVE',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (server_id, public_key),
    UNIQUE (server_id, assigned_ip)
);

CREATE INDEX idx_vpn_peers_user_id   ON vpn_peers (user_id);
CREATE INDEX idx_vpn_peers_device_id ON vpn_peers (device_id);
CREATE INDEX idx_vpn_peers_server_id ON vpn_peers (server_id);
CREATE INDEX idx_vpn_peers_status    ON vpn_peers (status);

ALTER TABLE ip_allocations ADD CONSTRAINT fk_ip_allocations_peer
    FOREIGN KEY (peer_id) REFERENCES vpn_peers (id) ON DELETE SET NULL;
