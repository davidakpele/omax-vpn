CREATE TYPE server_status AS ENUM ('HEALTHY', 'DEGRADED', 'MAINTENANCE', 'OFFLINE');

CREATE TABLE vpn_regions (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code       TEXT        NOT NULL UNIQUE,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE vpn_servers (
    id                 UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    region_id          UUID          NOT NULL REFERENCES vpn_regions (id),
    name               TEXT          NOT NULL UNIQUE,
    hostname           TEXT          NOT NULL,
    public_ip          INET          NOT NULL,
    country            CHAR(2)       NOT NULL,
    city               TEXT          NOT NULL,
    provider           TEXT          NOT NULL,
    status             server_status NOT NULL DEFAULT 'OFFLINE',
    capacity           INTEGER       NOT NULL DEFAULT 500 CHECK (capacity > 0),
    active_connections INTEGER       NOT NULL DEFAULT 0   CHECK (active_connections >= 0),
    wireguard_port     INTEGER       NOT NULL DEFAULT 51820,
    public_key         TEXT          NOT NULL,
    created_at         TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vpn_servers_region_id ON vpn_servers (region_id);
CREATE INDEX idx_vpn_servers_status    ON vpn_servers (status);
CREATE INDEX idx_vpn_servers_country   ON vpn_servers (country);
