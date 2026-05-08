package secret

import (
	"context"
	"testing"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretServiceList(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, svc *Service)
		filter  domain.SecretType
		wantLen int
	}{
		{
			name:    "empty store",
			setup:   func(t *testing.T, svc *Service) {},
			wantLen: 0,
		},
		{
			name: "all types",
			setup: func(t *testing.T, svc *Service) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
				require.NoError(t, svc.AddText(context.Background(), testKey, "note", "text", ""))
			},
			wantLen: 2,
		},
		{
			name: "filter by credential type",
			setup: func(t *testing.T, svc *Service) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
				require.NoError(t, svc.AddText(context.Background(), testKey, "note", "text", ""))
			},
			filter:  domain.SecretTypeCredential,
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, offlineMock())
			tt.setup(t, svc)

			list, err := svc.List(context.Background(), tt.filter)
			require.NoError(t, err)
			assert.Len(t, list, tt.wantLen)
		})
	}
}

func TestSecretServiceHasPending(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, svc *Service)
		wantPending bool
	}{
		{
			name:        "empty store",
			setup:       func(t *testing.T, svc *Service) {},
			wantPending: false,
		},
		{
			name: "after offline add",
			setup: func(t *testing.T, svc *Service) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
			},
			wantPending: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, offlineMock())
			tt.setup(t, svc)

			has, err := svc.HasPending(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.wantPending, has)
		})
	}
}
