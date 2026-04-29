package auth

import (
	"context"
	"errors"
	"testing"

	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthServiceLogin(t *testing.T) {
	tests := []struct {
		name       string
		resp       *pb.LoginResponse
		grpcErr    error
		wantErr    bool
		wantKeyLen int
	}{
		{
			name:       "success — returns 32-byte master key",
			resp:       validLoginResp("jwt.tok.en"),
			wantKeyLen: 32,
		},
		{
			name:    "server error",
			grpcErr: errors.New("unauthorized"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, &mockAuthGRPC{loginResp: tt.resp, loginErr: tt.grpcErr})
			key, err := svc.Login(context.Background(), "user@example.com", "password")
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, key, tt.wantKeyLen)
		})
	}
}
