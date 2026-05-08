package secret

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCmd_Empty(t *testing.T) {
	app, _, _ := newTestApp(t)

	cmd := NewListCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), "No secrets found")
}

func TestListCmd_WithSecrets(t *testing.T) {
	app, _, _ := newTestApp(t)
	addTestCredential(t, app, "github")

	cmd := NewListCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), "github")
	assert.Contains(t, out.String(), "credential")
}

func TestListCmd_FilterByType(t *testing.T) {
	app, _, _ := newTestApp(t)
	addTestCredential(t, app, "gh")

	masterKey, err := app.AuthSvc.DeriveMasterKey(context.Background(), testMasterPwd)
	require.NoError(t, err)
	require.NoError(t, app.SecretSvc.AddText(context.Background(), masterKey, "note", "hello", ""))

	cmd := NewListCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--type", "credential"})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), "gh")
	assert.NotContains(t, out.String(), "note")
}
