package secret

import (
	"bytes"
	"context"
	"errors"
	"testing"

	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncCmd_Success(t *testing.T) {
	app, _, secretsMock := newTestApp(t)
	loginApp(t, app)

	// server returns empty list → sync succeeds
	secretsMock.listResp = &pb.ListSecretsResponse{}
	secretsMock.listErr = nil
	secretsMock.createErr = nil

	cmd := NewSyncCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--master-password", testMasterPwd})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), "Sync completed")
}

func TestSyncCmd_PushFails(t *testing.T) {
	app, _, secretsMock := newTestApp(t)
	loginApp(t, app)

	// add pending credential, then make server return error on create (already default)
	masterKey, err := app.AuthSvc.DeriveMasterKey(context.Background(), testMasterPwd)
	require.NoError(t, err)
	require.NoError(t, app.SecretSvc.AddCredential(context.Background(), masterKey, "gh", "alice", "pass", "", ""))

	// pull also fails
	secretsMock.listErr = errors.New("unavailable")

	cmd := NewSyncCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--master-password", testMasterPwd})

	// push failures are silently skipped; pull error causes command to fail
	require.Error(t, cmd.ExecuteContext(context.Background()))
}
