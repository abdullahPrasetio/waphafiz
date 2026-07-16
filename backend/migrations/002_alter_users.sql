-- +migrate Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS family_group_id UUID        REFERENCES family_groups(id),
    ADD COLUMN IF NOT EXISTS role            VARCHAR(20) NOT NULL DEFAULT 'member',
    ADD COLUMN IF NOT EXISTS is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS last_seen_at    TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_users_family_group_id ON users(family_group_id);

-- +migrate Down
ALTER TABLE users
    DROP COLUMN IF EXISTS family_group_id,
    DROP COLUMN IF EXISTS role,
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS last_seen_at;
