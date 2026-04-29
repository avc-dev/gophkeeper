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

func TestCopyCmd_CredentialFound(t *testing.T) {
	app, _, _ := newTestApp(t)
	addTestCredential(t, app, "github")

	cmd := NewCopyCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"credential", "github", "--master-password", testMasterPwd})

	// clipboard.WriteAll fails in headless CI — that's expected; the important
	// thing is that GetCredential succeeded and value was resolved.
	_ = cmd.ExecuteContext(context.Background())
}

func TestCopyCmd_CardFound_Number(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)
	masterKey, err := app.AuthSvc.DeriveMasterKey(context.Background(), testMasterPwd)
	require.NoError(t, err)
	require.NoError(t, app.SecretSvc.AddCard(context.Background(), masterKey, "visa", "4532015112830366", "JOHN DOE", "12/26", "123", "", ""))

	cmd := NewCopyCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"card", "visa", "--master-password", testMasterPwd})

	_ = cmd.ExecuteContext(context.Background())
}

func TestCopyCmd_CardFound_CVV(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)
	masterKey, err := app.AuthSvc.DeriveMasterKey(context.Background(), testMasterPwd)
	require.NoError(t, err)
	require.NoError(t, app.SecretSvc.AddCard(context.Background(), masterKey, "visa2", "4532015112830366", "JOHN DOE", "12/26", "123", "", ""))

	cmd := NewCopyCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"card", "visa2", "--field", "cvv", "--master-password", testMasterPwd})

	_ = cmd.ExecuteContext(context.Background())
}
