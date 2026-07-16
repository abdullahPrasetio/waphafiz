-- +migrate Up
CREATE TABLE IF NOT EXISTS family_groups (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(200) NOT NULL,
    invite_code VARCHAR(50)  NOT NULL UNIQUE,
    created_by  UUID         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_family_groups_invite_code ON family_groups(invite_code);
CREATE INDEX IF NOT EXISTS idx_family_groups_deleted_at  ON family_groups(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS family_groups;
