package storage

import (
	"context"
	"testing"
	"time"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openTestDB открывает in-memory SQLite БД для тестов.
func openTestDB(t *testing.T) *authSecretDB {
	t.Helper()
	db, err := Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return &authSecretDB{
		auth:    NewAuthStorage(db),
		secrets: NewSecretStorage(db),
	}
}

type authSecretDB struct {
	auth    *AuthStorage
	secrets *SecretStorage
}

func newLocalSecret(name string, typ domain.SecretType) *LocalSecret {
	return &LocalSecret{
		Secret: domain.Secret{
			Type:    typ,
			Name:    name,
			Payload: []byte("encrypted-payload"),
		},
		SyncStatus: SyncStatusPending,
	}
}

// ─── AuthStorage ────────────────────────────────────────────────────────────

func TestAuthStorageSetGet(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "jwt token", key: "jwt_token", value: "eyJ..."},
		{name: "kdf salt", key: "kdf_salt", value: "base64salt=="},
		{name: "last sync", key: "last_sync_at", value: "2024-01-01T00:00:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			ctx := context.Background()

			require.NoError(t, db.auth.Set(ctx, tt.key, tt.value))
			got, err := db.auth.Get(ctx, tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.value, got)
		})
	}
}

func TestAuthStorageOverwrite(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	require.NoError(t, db.auth.Set(ctx, "key", "old"))
	require.NoError(t, db.auth.Set(ctx, "key", "new"))

	got, err := db.auth.Get(ctx, "key")
	require.NoError(t, err)
	assert.Equal(t, "new", got)
}

func TestAuthStorageMissingKey(t *testing.T) {
	db := openTestDB(t)
	got, err := db.auth.Get(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, "", got)
}

func TestAuthStorageDelete(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	require.NoError(t, db.auth.Set(ctx, "key", "value"))
	require.NoError(t, db.auth.Delete(ctx, "key"))

	got, err := db.auth.Get(ctx, "key")
	require.NoError(t, err)
	assert.Equal(t, "", got)
}

// ─── SecretStorage ───────────────────────────────────────────────────────────

func TestSecretCreate(t *testing.T) {
	tests := []struct {
		name string
		sec  *LocalSecret
	}{
		{
			name: "credential",
			sec:  newLocalSecret("github", domain.SecretTypeCredential),
		},
		{
			name: "card",
			sec:  newLocalSecret("visa", domain.SecretTypeCard),
		},
		{
			name: "с заданным id",
			sec: &LocalSecret{
				Secret: domain.Secret{
					ID:      uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001"),
					Type:    domain.SecretTypeText,
					Name:    "note",
					Payload: []byte("enc"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			got, err := db.secrets.Create(context.Background(), tt.sec)
			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, got.ID)
			assert.Equal(t, SyncStatusPending, got.SyncStatus)
			assert.Equal(t, int64(1), got.LocalVersion)
		})
	}
}

func TestSecretGet(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	created, err := db.secrets.Create(ctx, newLocalSecret("github", domain.SecretTypeCredential))
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{name: "существующий", id: created.ID},
		{name: "не найден", id: uuid.New(), wantErr: domain.ErrSecretNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.secrets.Get(ctx, tt.id)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, created.ID, got.ID)
			assert.Equal(t, "github", got.Name)
		})
	}
}

func TestSecretGetByName(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	_, err := db.secrets.Create(ctx, newLocalSecret("github", domain.SecretTypeCredential))
	require.NoError(t, err)

	tests := []struct {
		name    string
		secName string
		typ     domain.SecretType
		wantErr error
	}{
		{name: "найден", secName: "github", typ: domain.SecretTypeCredential},
		{name: "неверный тип", secName: "github", typ: domain.SecretTypeCard, wantErr: domain.ErrSecretNotFound},
		{name: "не существует", secName: "nonexistent", typ: domain.SecretTypeCredential, wantErr: domain.ErrSecretNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.secrets.GetByName(ctx, tt.secName, tt.typ)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.secName, got.Name)
		})
	}
}

func TestSecretList(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	_, err := db.secrets.Create(ctx, newLocalSecret("github", domain.SecretTypeCredential))
	require.NoError(t, err)
	_, err = db.secrets.Create(ctx, newLocalSecret("visa", domain.SecretTypeCard))
	require.NoError(t, err)

	tests := []struct {
		name    string
		typ     domain.SecretType
		wantLen int
	}{
		{name: "все типы", typ: "", wantLen: 2},
		{name: "только credentials", typ: domain.SecretTypeCredential, wantLen: 1},
		{name: "только cards", typ: domain.SecretTypeCard, wantLen: 1},
		{name: "text — пусто", typ: domain.SecretTypeText, wantLen: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.secrets.List(ctx, tt.typ)
			require.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
		})
	}
}

func TestSecretUpdate(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	created, err := db.secrets.Create(ctx, newLocalSecret("github", domain.SecretTypeCredential))
	require.NoError(t, err)

	tests := []struct {
		name     string
		id       uuid.UUID
		payload  []byte
		metadata string
		wantErr  error
	}{
		{name: "успешно", id: created.ID, payload: []byte("new-enc"), metadata: `{"url":"https://github.com"}`},
		{name: "не найден", id: uuid.New(), payload: []byte("enc"), wantErr: domain.ErrSecretNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.secrets.Update(ctx, tt.id, tt.payload, tt.metadata)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.payload, got.Payload)
			assert.Equal(t, int64(2), got.LocalVersion) // version bump
			assert.Equal(t, SyncStatusPending, got.SyncStatus)
		})
	}
}

func TestSecretDelete(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	created, err := db.secrets.Create(ctx, newLocalSecret("github", domain.SecretTypeCredential))
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{name: "успешное удаление", id: created.ID},
		{name: "повторное удаление", id: created.ID, wantErr: domain.ErrSecretNotFound},
		{name: "не существует", id: uuid.New(), wantErr: domain.ErrSecretNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.secrets.Delete(ctx, tt.id)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			// после soft delete — Get должен вернуть ErrSecretNotFound.
			_, getErr := db.secrets.Get(ctx, tt.id)
			assert.ErrorIs(t, getErr, domain.ErrSecretNotFound)
		})
	}
}

func TestSecretMarkSynced(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	created, err := db.secrets.Create(ctx, newLocalSecret("github", domain.SecretTypeCredential))
	require.NoError(t, err)

	serverID := uuid.New()
	serverTime := time.Now().UTC().Truncate(time.Second)

	err = db.secrets.MarkSynced(ctx, created.ID, serverID, 1, serverTime)
	require.NoError(t, err)

	got, err := db.secrets.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, SyncStatusSynced, got.SyncStatus)
	assert.Equal(t, int64(1), got.ServerVersion)
	require.NotNil(t, got.ServerID)
	assert.Equal(t, serverID, *got.ServerID)
}

func TestSecretUpsert(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	serverID := uuid.New()
	localID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	sec := &LocalSecret{
		Secret: domain.Secret{
			ID:        localID,
			Type:      domain.SecretTypeText,
			Name:      "note",
			Payload:   []byte("from-server"),
			UpdatedAt: now,
		},
		ServerID:      &serverID,
		LocalVersion:  1,
		ServerVersion: 1,
		SyncStatus:    SyncStatusSynced,
	}

	// первый upsert — вставка.
	require.NoError(t, db.secrets.Upsert(ctx, sec))
	got, err := db.secrets.Get(ctx, localID)
	require.NoError(t, err)
	assert.Equal(t, "note", got.Name)
	assert.Equal(t, SyncStatusSynced, got.SyncStatus)

	// второй upsert — обновление (не pending, значит перезапишется).
	sec.Payload = []byte("updated-from-server")
	sec.ServerVersion = 2
	require.NoError(t, db.secrets.Upsert(ctx, sec))
	got, err = db.secrets.Get(ctx, localID)
	require.NoError(t, err)
	assert.Equal(t, []byte("updated-from-server"), got.Payload)
}

func TestSecretUpsertSkipsPending(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	// создаём локальный секрет (pending).
	created, err := db.secrets.Create(ctx, newLocalSecret("github", domain.SecretTypeCredential))
	require.NoError(t, err)
	assert.Equal(t, SyncStatusPending, created.SyncStatus)

	serverID := uuid.New()
	now := time.Now().UTC()

	// upsert с сервера НЕ должен перезаписать pending запись.
	incoming := &LocalSecret{
		Secret: domain.Secret{
			ID:        created.ID,
			Type:      domain.SecretTypeCredential,
			Name:      "github",
			Payload:   []byte("server-version"),
			UpdatedAt: now,
		},
		ServerID:      &serverID,
		LocalVersion:  1,
		ServerVersion: 1,
		SyncStatus:    SyncStatusSynced,
	}
	require.NoError(t, db.secrets.Upsert(ctx, incoming))

	got, err := db.secrets.Get(ctx, created.ID)
	require.NoError(t, err)
	// payload должен остаться от локальной версии.
	assert.Equal(t, []byte("encrypted-payload"), got.Payload)
	assert.Equal(t, SyncStatusPending, got.SyncStatus)
}

func TestSecretGetByServerID(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	created, err := db.secrets.Create(ctx, newLocalSecret("github", domain.SecretTypeCredential))
	require.NoError(t, err)

	serverID := uuid.New()
	require.NoError(t, db.secrets.MarkSynced(ctx, created.ID, serverID, 1, time.Now()))

	t.Run("found by server id", func(t *testing.T) {
		got, err := db.secrets.GetByServerID(ctx, serverID)
		require.NoError(t, err)
		assert.Equal(t, "github", got.Name)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := db.secrets.GetByServerID(ctx, uuid.New())
		assert.ErrorIs(t, err, domain.ErrSecretNotFound)
	})
}

func TestSecretPurge(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	created, err := db.secrets.Create(ctx, newLocalSecret("github", domain.SecretTypeCredential))
	require.NoError(t, err)

	serverID := uuid.New()
	require.NoError(t, db.secrets.MarkSynced(ctx, created.ID, serverID, 1, time.Now()))

	// мягкое удаление — помечает deleted=1, sync_status=pending.
	require.NoError(t, db.secrets.Delete(ctx, created.ID))

	// после Delete запись pending; Purge ожидает sync_status=synced, значит — no-op.
	require.NoError(t, db.secrets.Purge(ctx, created.ID))

	// вручную помечаем как synced для теста физического удаления.
	_, err = db.secrets.db.ExecContext(ctx,
		`UPDATE secrets SET sync_status = 'synced' WHERE id = ?`, created.ID.String())
	require.NoError(t, err)

	require.NoError(t, db.secrets.Purge(ctx, created.ID))

	// запись должна быть физически удалена.
	_, err = db.secrets.Get(ctx, created.ID)
	assert.ErrorIs(t, err, domain.ErrSecretNotFound)
}

func TestCheckpoint(t *testing.T) {
	// Checkpoint на :memory: — WAL не применяется, но код не должен упасть.
	db, err := Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = Checkpoint(context.Background(), db)
	require.NoError(t, err)
}

func TestListPending(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	sec1, err := db.secrets.Create(ctx, newLocalSecret("github", domain.SecretTypeCredential))
	require.NoError(t, err)
	_, err = db.secrets.Create(ctx, newLocalSecret("visa", domain.SecretTypeCard))
	require.NoError(t, err)

	// помечаем первый как synced.
	require.NoError(t, db.secrets.MarkSynced(ctx, sec1.ID, uuid.New(), 1, time.Now()))

	pending, err := db.secrets.ListPending(ctx)
	require.NoError(t, err)
	assert.Len(t, pending, 1)
	assert.Equal(t, "visa", pending[0].Name)
}
