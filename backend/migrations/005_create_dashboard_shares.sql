-- +migrate Up
CREATE TABLE IF NOT EXISTS dashboard_shares (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID         NOT NULL REFERENCES users(id),
    created_by     UUID         NOT NULL REFERENCES users(id),
    token_hash     VARCHAR(64)  NOT NULL UNIQUE,
    label          VARCHAR(100) NOT NULL,
    expires_at     TIMESTAMPTZ,
    revoked_at     TIMESTAMPTZ,
    last_viewed_at TIMESTAMPTZ,
    view_count     INT          NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dashboard_shares_user ON dashboard_shares(user_id);

-- +migrate Down
DROP TABLE IF EXISTS dashboard_shares;
