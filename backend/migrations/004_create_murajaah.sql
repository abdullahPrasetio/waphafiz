-- +migrate Up
CREATE TABLE IF NOT EXISTS murajaah_schedule (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID        NOT NULL REFERENCES users(id),
    type           VARCHAR(20) NOT NULL,
    surah_number   INT         NOT NULL,
    ayat_start     INT         NOT NULL,
    ayat_end       INT         NOT NULL,
    scheduled_date DATE        NOT NULL,
    completed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_murajaah_user_date ON murajaah_schedule(user_id, scheduled_date);

CREATE TABLE IF NOT EXISTS murajaah_log (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id  UUID        NOT NULL REFERENCES murajaah_schedule(id),
    user_id      UUID        NOT NULL REFERENCES users(id),
    completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes        VARCHAR(500)
);

CREATE INDEX IF NOT EXISTS idx_murajaah_log_schedule_id ON murajaah_log(schedule_id);
CREATE INDEX IF NOT EXISTS idx_murajaah_log_user_id     ON murajaah_log(user_id);

-- +migrate Down
DROP TABLE IF EXISTS murajaah_log;
DROP TABLE IF EXISTS murajaah_schedule;
