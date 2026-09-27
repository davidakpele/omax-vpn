CREATE TYPE subscription_status AS ENUM ('ACTIVE', 'CANCELLED', 'EXPIRED', 'PAST_DUE');
CREATE TYPE billing_interval  AS ENUM ('MONTHLY', 'YEARLY');

CREATE TABLE plans (
    id               UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT             NOT NULL UNIQUE,
    display_name     TEXT             NOT NULL,
    price_cents      INTEGER          NOT NULL CHECK (price_cents >= 0),
    currency         CHAR(3)          NOT NULL DEFAULT 'USD',
    billing_interval billing_interval NOT NULL,
    max_devices      INTEGER          NOT NULL DEFAULT 5,
    is_active        BOOLEAN          NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE subscriptions (
    id                   UUID                NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID                NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    plan_id              UUID                NOT NULL REFERENCES plans (id),
    status               subscription_status NOT NULL DEFAULT 'ACTIVE',
    current_period_start TIMESTAMPTZ         NOT NULL,
    current_period_end   TIMESTAMPTZ         NOT NULL,
    cancelled_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_subscriptions_active_user
    ON subscriptions (user_id)
    WHERE status = 'ACTIVE';

CREATE INDEX idx_subscriptions_user_id ON subscriptions (user_id);
CREATE INDEX idx_subscriptions_status  ON subscriptions (status);

CREATE TABLE payments (
    id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    subscription_id      UUID        REFERENCES subscriptions (id),
    amount_cents         INTEGER     NOT NULL CHECK (amount_cents >= 0),
    currency             CHAR(3)     NOT NULL DEFAULT 'USD',
    status               TEXT        NOT NULL,
    provider             TEXT        NOT NULL,
    provider_payment_id  TEXT        UNIQUE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_user_id         ON payments (user_id);
CREATE INDEX idx_payments_subscription_id ON payments (subscription_id);
