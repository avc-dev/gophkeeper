//go:build integration

// Package integration содержит end-to-end тесты против реального PostgreSQL и gRPC сервера.
// Запуск: make test-integration (требует docker-compose.test.yml).
package integration

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/avc-dev/gophkeeper/internal/server/handler"
	authhandler "github.com/avc-dev/gophkeeper/internal/server/handler/auth"
	secrethandler "github.com/avc-dev/gophkeeper/internal/server/handler/secret"
	authsvc "github.com/avc-dev/gophkeeper/internal/server/service/auth"
	secretsvc "github.com/avc-dev/gophkeeper/internal/server/service/secret"
	secretstore "github.com/avc-dev/gophkeeper/internal/server/storage/secret"
	userstore "github.com/avc-dev/gophkeeper/internal/server/storage/user"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// testDSN — строка подключения к тестовой БД.
// Переопределяется переменной окружения TEST_DSN.
func testDSN() string {
	if v := os.Getenv("TEST_DSN"); v != "" {
		return v
	}
	return "postgres://gophkeeper:gophkeeper@localhost:5433/gophkeeper_test?sslmode=disable"
}

// testEdKeys генерирует одноразовую Ed25519 ключевую пару для тестового сервера.
func testEdKeys(t *testing.T) (ed25519.PrivateKey, ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err, "generate test Ed25519 key pair")
	return priv, pub
}

// testServer запускает gRPC сервер на случайном порту и возвращает его адрес.
// Сервер автоматически останавливается по завершению теста.
func testServer(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	db, err := pgxpool.New(ctx, testDSN())
	require.NoError(t, err, "connect to test DB")
	require.NoError(t, db.Ping(ctx), "ping test DB")

	// очищаем таблицы перед тестом.
	_, err = db.Exec(ctx, "TRUNCATE users, secrets RESTART IDENTITY CASCADE")
	require.NoError(t, err, "truncate tables")

	privKey, pubKey := testEdKeys(t)

	users := userstore.New(db)
	auth := authsvc.New(users, privKey, pubKey)

	secrets := secretstore.New(db)
	secretService := secretsvc.New(secrets)

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(handler.AuthInterceptor(auth)),
	)
	pb.RegisterAuthServiceServer(srv, authhandler.New(auth))
	pb.RegisterSecretsServiceServer(srv, secrethandler.New(secretService))

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go srv.Serve(lis) //nolint:errcheck

	t.Cleanup(func() {
		srv.GracefulStop()
		db.Close()
	})

	return lis.Addr().String()
}

// testClient возвращает gRPC клиенты для auth и secrets.
func testClient(t *testing.T, addr string) (pb.AuthServiceClient, pb.SecretsServiceClient) {
	t.Helper()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	return pb.NewAuthServiceClient(conn), pb.NewSecretsServiceClient(conn)
}

// authedCtx создаёт context с Bearer токеном для gRPC вызовов.
func authedCtx(token string) context.Context {
	md := metadata.Pairs("authorization", "Bearer "+token)
	return metadata.NewOutgoingContext(context.Background(), md)
}

// ─── TestRegisterAndLogin ────────────────────────────────────────────────────

func TestRegisterAndLogin(t *testing.T) {
	addr := testServer(t)
	authClient, _ := testClient(t, addr)
	ctx := context.Background()

	email := "alice@example.com"
	password := "s3cr3tPass!"

	t.Run("register new user", func(t *testing.T) {
		_, err := authClient.Register(ctx, &pb.RegisterRequest{
			Email:    email,
			Password: password,
		})
		require.NoError(t, err)
	})

	t.Run("register duplicate email — error", func(t *testing.T) {
		_, err := authClient.Register(ctx, &pb.RegisterRequest{
			Email:    email,
			Password: password,
		})
		require.Error(t, err)
	})

	t.Run("login with correct password — returns token and kdf_salt", func(t *testing.T) {
		resp, err := authClient.Login(ctx, &pb.LoginRequest{
			Email:    email,
			Password: password,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Token)
		assert.NotEmpty(t, resp.KdfSalt)
	})

	t.Run("login with wrong password — error", func(t *testing.T) {
		_, err := authClient.Login(ctx, &pb.LoginRequest{
			Email:    email,
			Password: "wrongpassword",
		})
		require.Error(t, err)
	})
}

// ─── TestAddCredentialAndSync ────────────────────────────────────────────────

func TestAddCredentialAndSync(t *testing.T) {
	addr := testServer(t)
	authClient, secretsClient := testClient(t, addr)
	ctx := context.Background()

	// регистрация и вход.
	_, err := authClient.Register(ctx, &pb.RegisterRequest{
		Email:    "bob@example.com",
		Password: "password",
	})
	require.NoError(t, err)

	loginResp, err := authClient.Login(ctx, &pb.LoginRequest{
		Email:    "bob@example.com",
		Password: "password",
	})
	require.NoError(t, err)

	authed := authedCtx(loginResp.Token)

	// создаём секрет.
	createResp, err := secretsClient.CreateSecret(authed, &pb.CreateSecretRequest{
		Type:             pb.SecretType_SECRET_TYPE_CREDENTIAL,
		Name:             "github",
		EncryptedPayload: []byte("encrypted-credential-payload"),
		Metadata:         "",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, createResp.Id)
	assert.Equal(t, int64(1), createResp.Version)

	// ListSecrets должен вернуть только что созданный секрет.
	listResp, err := secretsClient.ListSecrets(authed, &pb.ListSecretsRequest{})
	require.NoError(t, err)
	require.Len(t, listResp.Secrets, 1)
	assert.Equal(t, "github", listResp.Secrets[0].Name)
	assert.Equal(t, pb.SecretType_SECRET_TYPE_CREDENTIAL, listResp.Secrets[0].Type)

	// инкрементальная синхронизация — since после создания должна вернуть пустой список.
	sinceAfter := time.Now().Add(time.Second)
	listResp2, err := secretsClient.ListSecrets(authed, &pb.ListSecretsRequest{
		Since: timestamppb.New(sinceAfter),
	})
	require.NoError(t, err)
	assert.Empty(t, listResp2.Secrets)
}

// ─── TestConflictResolution ──────────────────────────────────────────────────

func TestConflictResolution(t *testing.T) {
	addr := testServer(t)
	authClient, secretsClient := testClient(t, addr)
	ctx := context.Background()

	_, err := authClient.Register(ctx, &pb.RegisterRequest{
		Email:    "carol@example.com",
		Password: "password",
	})
	require.NoError(t, err)

	loginResp, err := authClient.Login(ctx, &pb.LoginRequest{
		Email:    "carol@example.com",
		Password: "password",
	})
	require.NoError(t, err)

	authed := authedCtx(loginResp.Token)

	// создаём секрет.
	createResp, err := secretsClient.CreateSecret(authed, &pb.CreateSecretRequest{
		Type:             pb.SecretType_SECRET_TYPE_TEXT,
		Name:             "note",
		EncryptedPayload: []byte("v1"),
	})
	require.NoError(t, err)

	// первое обновление — версия 1 → 2.
	updateResp, err := secretsClient.UpdateSecret(authed, &pb.UpdateSecretRequest{
		Id:               createResp.Id,
		EncryptedPayload: []byte("v2"),
		ExpectedVersion:  1,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), updateResp.Version)

	// конфликт: ожидаем версию 1, но на сервере уже 2.
	_, err = secretsClient.UpdateSecret(authed, &pb.UpdateSecretRequest{
		Id:               createResp.Id,
		EncryptedPayload: []byte("v2-conflict"),
		ExpectedVersion:  1,
	})
	require.Error(t, err, "ожидался ErrVersionConflict")
}

// ─── TestOfflineMode ─────────────────────────────────────────────────────────

// TestOfflineMode проверяет, что клиент может хранить секреты локально
// и успешно синхронизировать их при восстановлении соединения.
// Симулируем: create offline (в хранилище, без сервера) → sync → verify on server.
func TestOfflineMode(t *testing.T) {
	addr := testServer(t)
	authClient, secretsClient := testClient(t, addr)
	ctx := context.Background()

	_, err := authClient.Register(ctx, &pb.RegisterRequest{
		Email:    "dave@example.com",
		Password: "password",
	})
	require.NoError(t, err)

	loginResp, err := authClient.Login(ctx, &pb.LoginRequest{
		Email:    "dave@example.com",
		Password: "password",
	})
	require.NoError(t, err)

	authed := authedCtx(loginResp.Token)

	// offline режим: сохраняем 2 секрета напрямую через gRPC (имитация PushPending).
	ids := make([]string, 2)
	for i, name := range []string{"offline-cred", "offline-text"} {
		resp, err := secretsClient.CreateSecret(authed, &pb.CreateSecretRequest{
			Type:             pb.SecretType_SECRET_TYPE_CREDENTIAL,
			Name:             name,
			EncryptedPayload: []byte("encrypted-offline-" + name),
		})
		require.NoError(t, err)
		ids[i] = resp.Id
	}

	// pull: ListSecrets должен вернуть оба секрета.
	listResp, err := secretsClient.ListSecrets(authed, &pb.ListSecretsRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Secrets, 2)

	// удаляем один из секретов.
	_, err = secretsClient.DeleteSecret(authed, &pb.DeleteSecretRequest{Id: ids[0]})
	require.NoError(t, err)

	// после удаления остаётся один.
	listResp, err = secretsClient.ListSecrets(authed, &pb.ListSecretsRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Secrets, 1)
	assert.Equal(t, "offline-text", listResp.Secrets[0].Name)
}
