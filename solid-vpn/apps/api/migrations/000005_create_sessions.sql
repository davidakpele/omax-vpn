CREATE TYPE session_status AS ENUM ('ACTIVE', 'ENDED', 'EXPIRED', 'ERROR');

CREATE TABLE vpn_sessions (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID           NOT NULL REFERENCES users (id)       ON DELETE CASCADE,
    device_id       UUID           NOT NULL REFERENCES devices (id)     ON DELETE CASCADE,
    server_id       UUID           NOT NULL REFERENCES vpn_servers (id) ON DELETE CASCADE,
    peer_id         UUID           REFERENCES vpn_peers (id)            ON DELETE SET NULL,
    status          session_status NOT NULL DEFAULT 'ACTIVE',
    bytes_sent      BIGINT         NOT NULL DEFAULT 0,
    bytes_received  BIGINT         NOT NULL DEFAULT 0,
    client_ip       INET,
    started_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    ended_at        TIMESTAMPTZ,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vpn_sessions_user_id   ON vpn_sessions (user_id);
CREATE INDEX idx_vpn_sessions_device_id ON vpn_sessions (device_id);
CREATE INDEX idx_vpn_sessions_server_id ON vpn_sessions (server_id);
CREATE INDEX idx_vpn_sessions_status    ON vpn_sessions (status);
CREATE INDEX idx_vpn_sessions_started_at ON vpn_sessions (started_at DESC);
