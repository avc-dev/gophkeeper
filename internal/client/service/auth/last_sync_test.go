package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthServiceLastSyncAt(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	later := now.Add(time.Hour)

	tests := []struct {
		name     string
		setTime  *time.Time
		wantNil  bool
		wantTime *time.Time
	}{
		{name: "nil before first sync", wantNil: true},
		{name: "stores and retrieves time", setTime: &now, wantTime: &now},
		{name: "overwrite updates time", setTime: &later, wantTime: &later},
	}

	// одна БД на все кейсы: накопительная проверка
	svc := newTestService(t, &mockAuthGRPC{})
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setTime != nil {
				require.NoError(t, svc.SetLastSyncAt(ctx, *tt.setTime))
			}

			got, err := svc.GetLastSyncAt(ctx)
			require.NoError(t, err)

			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.WithinDuration(t, *tt.wantTime, *got, time.Second)
		})
	}
}
