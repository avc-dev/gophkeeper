CREATE TABLE secrets (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       TEXT        NOT NULL,
    name       TEXT        NOT NULL,
    payload    BYTEA       NOT NULL,
    metadata   TEXT        NOT NULL DEFAULT '',
    version    BIGINT      NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- индекс для выборки секретов пользователя + инкрементальной синхронизации по updated_at
CREATE INDEX secrets_user_updated_idx ON secrets(user_id, updated_at);
