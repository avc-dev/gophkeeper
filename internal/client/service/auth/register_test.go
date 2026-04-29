package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthServiceRegister(t *testing.T) {
	tests := []struct {
		name    string
		grpcErr error
		wantErr bool
	}{
		{name: "success"},
		{name: "server unavailable", grpcErr: errors.New("unavailable"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, &mockAuthGRPC{registerErr: tt.grpcErr})
			err := svc.Register(context.Background(), "user@example.com", "password")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
