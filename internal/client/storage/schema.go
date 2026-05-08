package storage

import (
	"database/sql"
	"fmt"
)

// schema содержит DDL для локальной БД клиента.
const schema = `
CREATE TABLE IF NOT EXISTS auth (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS secrets (
    id                TEXT PRIMARY KEY,
    server_id         TEXT,
    type              TEXT NOT NULL,
    name              TEXT NOT NULL,
    payload           BLOB NOT NULL,
    metadata          TEXT NOT NULL DEFAULT '',
    local_version     INTEGER NOT NULL DEFAULT 1,
    server_version    INTEGER NOT NULL DEFAULT 0,
    updated_at        TEXT NOT NULL,
    server_updated_at TEXT,
    sync_status       TEXT NOT NULL DEFAULT 'pending',
    deleted           INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS secrets_sync_idx  ON secrets(sync_status);
CREATE INDEX IF NOT EXISTS secrets_name_idx  ON secrets(name, type);
`

// applySchema создаёт таблицы если они не существуют.
func applySchema(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
