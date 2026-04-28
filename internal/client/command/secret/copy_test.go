package secret

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopyCmd_UnsupportedType(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)

	cmd := NewCopyCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"text", "note", "--master-password", testMasterPwd})

	err := cmd.ExecuteContext(context.Background())
	require.Error(t, err)
}

func TestCopyCmd_CredentialNotFound(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)

	cmd := NewCopyCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"credential", "missing", "--master-password", testMasterPwd})

	require.Error(t, cmd.ExecuteContext(context.Background()))
}

func TestCopyCmd_CardNotFound(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)

	cmd := NewCopyCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"card", "missing", "--master-password", testMasterPwd})

	require.Error(t, cmd.ExecuteContext(context.Background()))
}
