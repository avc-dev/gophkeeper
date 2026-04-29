package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ─── mock tokenValidator ──────────────────────────────────────────────────────

type mockValidator struct {
	userID uuid.UUID
	err    error
}

func (m *mockValidator) ValidateToken(_ string) (uuid.UUID, error) {
	return m.userID, m.err
}

// ─── helper ───────────────────────────────────────────────────────────────────

// captureHandler returns a grpc.UnaryHandler that stores the received context.
func captureHandler(captured *context.Context) grpc.UnaryHandler {
	return func(ctx context.Context, req any) (any, error) {
		*captured = ctx
		return "ok", nil
	}
}

func authedCtx(token string) context.Context {
	md := metadata.Pairs("authorization", "Bearer "+token)
	return metadata.NewIncomingContext(context.Background(), md)
}

// ─── AuthInterceptor ──────────────────────────────────────────────────────────

func TestAuthInterceptor_PublicMethod_NoToken(t *testing.T) {
	v := &mockValidator{}
	interceptor := AuthInterceptor(v)
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.AuthService/Login"}

	resp, err := interceptor(context.Background(), nil, info, func(ctx context.Context, req any) (any, error) {
		return "passed", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "passed", resp)
}

func TestAuthInterceptor_PublicMethod_Register(t *testing.T) {
	interceptor := AuthInterceptor(&mockValidator{})
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.AuthService/Register"}

	_, err := interceptor(context.Background(), nil, info, func(ctx context.Context, req any) (any, error) {
		return nil, nil
	})
	require.NoError(t, err)
}

func TestAuthInterceptor_PublicMethod_Ping(t *testing.T) {
	interceptor := AuthInterceptor(&mockValidator{})
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.SecretsService/Ping"}

	_, err := interceptor(context.Background(), nil, info, func(ctx context.Context, req any) (any, error) {
		return nil, nil
	})
	require.NoError(t, err)
}

func TestAuthInterceptor_MissingMetadata(t *testing.T) {
	interceptor := AuthInterceptor(&mockValidator{})
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.SecretsService/CreateSecret"}

	_, err := interceptor(context.Background(), nil, info, captureHandler(new(context.Context)))

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestAuthInterceptor_MissingToken(t *testing.T) {
	interceptor := AuthInterceptor(&mockValidator{})
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.SecretsService/CreateSecret"}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})
	_, err := interceptor(ctx, nil, info, captureHandler(new(context.Context)))

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestAuthInterceptor_InvalidToken(t *testing.T) {
	v := &mockValidator{err: errors.New("bad signature")}
	interceptor := AuthInterceptor(v)
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.SecretsService/CreateSecret"}

	_, err := interceptor(authedCtx("bad.token"), nil, info, captureHandler(new(context.Context)))

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestAuthInterceptor_ValidToken_InjectsUserID(t *testing.T) {
	wantID := uuid.New()
	v := &mockValidator{userID: wantID}
	interceptor := AuthInterceptor(v)
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.SecretsService/CreateSecret"}

	var captured context.Context
	_, err := interceptor(authedCtx("valid.token"), nil, info, captureHandler(&captured))

	require.NoError(t, err)
	gotID, ok := UserIDFromContext(captured)
	require.True(t, ok)
	assert.Equal(t, wantID, gotID)
}

// ─── UserIDFromContext / WithUserID ───────────────────────────────────────────

func TestUserIDFromContext_NotPresent(t *testing.T) {
	_, ok := UserIDFromContext(context.Background())
	assert.False(t, ok)
}

func TestWithUserID_RoundTrip(t *testing.T) {
	id := uuid.New()
	ctx := WithUserID(context.Background(), id)
	got, ok := UserIDFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, id, got)
}
