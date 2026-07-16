-- +migrate Up
CREATE TABLE IF NOT EXISTS hafalan_progress (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id),
    surah_number INT         NOT NULL,
    ayat_start   INT         NOT NULL,
    ayat_end     INT         NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'belum',
    noted_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_hafalan_user_id      ON hafalan_progress(user_id);
CREATE INDEX IF NOT EXISTS idx_hafalan_surah        ON hafalan_progress(user_id, surah_number);
CREATE INDEX IF NOT EXISTS idx_hafalan_noted_at     ON hafalan_progress(user_id, noted_at);
CREATE INDEX IF NOT EXISTS idx_hafalan_deleted_at   ON hafalan_progress(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS hafalan_progress;
