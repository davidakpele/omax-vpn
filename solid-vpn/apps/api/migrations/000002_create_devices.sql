CREATE TYPE device_status AS ENUM ('ACTIVE', 'REVOKED');

CREATE TABLE devices (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        TEXT          NOT NULL,
    public_key  TEXT          NOT NULL UNIQUE,
    status      device_status NOT NULL DEFAULT 'ACTIVE',
    last_seen   TIMESTAMPTZ,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_devices_user_id    ON devices (user_id);
CREATE INDEX idx_devices_public_key ON devices (public_key);
CREATE INDEX idx_devices_status     ON devices (status);
