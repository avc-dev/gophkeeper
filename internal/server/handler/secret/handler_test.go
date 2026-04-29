package secret

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/avc-dev/gophkeeper/internal/server/handler"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockService — заглушка сервиса для изолированного тестирования хендлера.
type mockService struct {
	secret  *domain.Secret
	secrets []*domain.Secret
	err     error
}

func (m *mockService) Create(_ context.Context, _ uuid.UUID, _ *domain.Secret) (*domain.Secret, error) {
	return m.secret, m.err
}

func (m *mockService) Get(_ context.Context, _, _ uuid.UUID) (*domain.Secret, error) {
	return m.secret, m.err
}

func (m *mockService) List(_ context.Context, _ uuid.UUID, _ *time.Time) ([]*domain.Secret, error) {
	return m.secrets, m.err
}

func (m *mockService) Update(_ context.Context, _, _ uuid.UUID, _ []byte, _ string, _ int64) (*domain.Secret, error) {
	return m.secret, m.err
}

func (m *mockService) Delete(_ context.Context, _, _ uuid.UUID) error {
	return m.err
}

// фиксированные идентификаторы для детерминированных тестов.
var (
	testUserID   = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	testSecretID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

// ctxWithUser возвращает контекст с user_id (имитирует AuthInterceptor).
func ctxWithUser() context.Context {
	return handler.WithUserID(context.Background(), testUserID)
}

func testSecret() *domain.Secret {
	return &domain.Secret{
		ID:       testSecretID,
		UserID:   testUserID,
		Type:     domain.SecretTypeCredential,
		Name:     "github",
		Payload:  []byte("encrypted"),
		Metadata: `{"url":"https://github.com"}`,
		Version:  1,
	}
}

func TestHandlerCreateSecret(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		req      *pb.CreateSecretRequest
		mockErr  error
		wantCode codes.Code
	}{
		{
			name:     "успешное создание",
			ctx:      ctxWithUser(),
			req:      &pb.CreateSecretRequest{Name: "github", EncryptedPayload: []byte("enc"), Type: pb.SecretType_SECRET_TYPE_CREDENTIAL},
			wantCode: codes.OK,
		},
		{
			name:     "пустое имя",
			ctx:      ctxWithUser(),
			req:      &pb.CreateSecretRequest{Name: "", EncryptedPayload: []byte("enc")},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "nil payload",
			ctx:      ctxWithUser(),
			req:      &pb.CreateSecretRequest{Name: "github", EncryptedPayload: nil},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "нет user_id в контексте",
			ctx:      context.Background(),
			req:      &pb.CreateSecretRequest{Name: "github", EncryptedPayload: []byte("enc")},
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "ошибка сервиса",
			ctx:      ctxWithUser(),
			req:      &pb.CreateSecretRequest{Name: "github", EncryptedPayload: []byte("enc")},
			mockErr:  errors.New("db error"),
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(&mockService{secret: testSecret(), err: tt.mockErr})
			resp, err := h.CreateSecret(tt.ctx, tt.req)

			if tt.wantCode == codes.OK {
				require.NoError(t, err)
				assert.Equal(t, testSecretID.String(), resp.Id)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.wantCode, status.Code(err))
			}
		})
	}
}

func TestHandlerGetSecret(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		id       string
		mockErr  error
		wantCode codes.Code
	}{
		{
			name:     "успешное получение",
			ctx:      ctxWithUser(),
			id:       testSecretID.String(),
			wantCode: codes.OK,
		},
		{
			name:     "нет user_id в контексте",
			ctx:      context.Background(),
			id:       testSecretID.String(),
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "невалидный id",
			ctx:      ctxWithUser(),
			id:       "not-a-uuid",
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "секрет не найден",
			ctx:      ctxWithUser(),
			id:       testSecretID.String(),
			mockErr:  domain.ErrSecretNotFound,
			wantCode: codes.NotFound,
		},
		{
			name:     "ошибка сервиса",
			ctx:      ctxWithUser(),
			id:       testSecretID.String(),
			mockErr:  errors.New("db error"),
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(&mockService{secret: testSecret(), err: tt.mockErr})
			resp, err := h.GetSecret(tt.ctx, &pb.GetSecretRequest{Id: tt.id})

			if tt.wantCode == codes.OK {
				require.NoError(t, err)
				assert.Equal(t, testSecretID.String(), resp.Secret.Id)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.wantCode, status.Code(err))
			}
		})
	}
}

func TestHandlerListSecrets(t *testing.T) {
	twoSecrets := []*domain.Secret{testSecret(), testSecret()}

	tests := []struct {
		name     string
		ctx      context.Context
		req      *pb.ListSecretsRequest
		mockSvc  *mockService
		wantCode codes.Code
		wantLen  int
	}{
		{
			name:     "все секреты",
			ctx:      ctxWithUser(),
			req:      &pb.ListSecretsRequest{},
			mockSvc:  &mockService{secrets: twoSecrets},
			wantCode: codes.OK,
			wantLen:  2,
		},
		{
			name:     "нет user_id в контексте",
			ctx:      context.Background(),
			req:      &pb.ListSecretsRequest{},
			mockSvc:  &mockService{},
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "ошибка сервиса",
			ctx:      ctxWithUser(),
			req:      &pb.ListSecretsRequest{},
			mockSvc:  &mockService{err: errors.New("db error")},
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(tt.mockSvc)
			resp, err := h.ListSecrets(tt.ctx, tt.req)

			if tt.wantCode == codes.OK {
				require.NoError(t, err)
				assert.Len(t, resp.Secrets, tt.wantLen)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.wantCode, status.Code(err))
			}
		})
	}
}

func TestHandlerUpdateSecret(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		req      *pb.UpdateSecretRequest
		mockErr  error
		wantCode codes.Code
	}{
		{
			name:     "успешное обновление",
			ctx:      ctxWithUser(),
			req:      &pb.UpdateSecretRequest{Id: testSecretID.String(), EncryptedPayload: []byte("new-enc"), ExpectedVersion: 1},
			wantCode: codes.OK,
		},
		{
			name:     "nil payload",
			ctx:      ctxWithUser(),
			req:      &pb.UpdateSecretRequest{Id: testSecretID.String(), EncryptedPayload: nil},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "нет user_id в контексте",
			ctx:      context.Background(),
			req:      &pb.UpdateSecretRequest{Id: testSecretID.String(), EncryptedPayload: []byte("enc")},
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "невалидный id",
			ctx:      ctxWithUser(),
			req:      &pb.UpdateSecretRequest{Id: "bad", EncryptedPayload: []byte("enc")},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "конфликт версий",
			ctx:      ctxWithUser(),
			req:      &pb.UpdateSecretRequest{Id: testSecretID.String(), EncryptedPayload: []byte("enc"), ExpectedVersion: 99},
			mockErr:  domain.ErrVersionConflict,
			wantCode: codes.Aborted,
		},
		{
			name:     "секрет не найден",
			ctx:      ctxWithUser(),
			req:      &pb.UpdateSecretRequest{Id: testSecretID.String(), EncryptedPayload: []byte("enc")},
			mockErr:  domain.ErrSecretNotFound,
			wantCode: codes.NotFound,
		},
		{
			name:     "ошибка сервиса",
			ctx:      ctxWithUser(),
			req:      &pb.UpdateSecretRequest{Id: testSecretID.String(), EncryptedPayload: []byte("enc")},
			mockErr:  errors.New("db error"),
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(&mockService{secret: testSecret(), err: tt.mockErr})
			resp, err := h.UpdateSecret(tt.ctx, tt.req)

			if tt.wantCode == codes.OK {
				require.NoError(t, err)
				assert.Equal(t, int64(1), resp.Version)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.wantCode, status.Code(err))
			}
		})
	}
}

func TestHandlerDeleteSecret(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		id       string
		mockErr  error
		wantCode codes.Code
	}{
		{
			name:     "успешное удаление",
			ctx:      ctxWithUser(),
			id:       testSecretID.String(),
			wantCode: codes.OK,
		},
		{
			name:     "нет user_id в контексте",
			ctx:      context.Background(),
			id:       testSecretID.String(),
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "невалидный id",
			ctx:      ctxWithUser(),
			id:       "bad-uuid",
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "секрет не найден",
			ctx:      ctxWithUser(),
			id:       testSecretID.String(),
			mockErr:  domain.ErrSecretNotFound,
			wantCode: codes.NotFound,
		},
		{
			name:     "ошибка сервиса",
			ctx:      ctxWithUser(),
			id:       testSecretID.String(),
			mockErr:  errors.New("db error"),
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(&mockService{err: tt.mockErr})
			resp, err := h.DeleteSecret(tt.ctx, &pb.DeleteSecretRequest{Id: tt.id})

			if tt.wantCode == codes.OK {
				require.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.wantCode, status.Code(err))
			}
		})
	}
}
