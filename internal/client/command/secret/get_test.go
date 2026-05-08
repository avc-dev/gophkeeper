package secret

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCredentialCmd(t *testing.T) {
	app, _, _ := newTestApp(t)
	addTestCredential(t, app, "github")

	cmd := NewGetCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"credential", "github", "--master-password", testMasterPwd})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), "alice")
	assert.Contains(t, out.String(), "s3cr3t")
}

func TestGetCredentialCmd_NotFound(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)

	cmd := NewGetCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"credential", "nonexistent", "--master-password", testMasterPwd})

	require.Error(t, cmd.ExecuteContext(context.Background()))
}

func TestGetCardCmd(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)
	masterKey, err := app.AuthSvc.DeriveMasterKey(context.Background(), testMasterPwd)
	require.NoError(t, err)
	require.NoError(t, app.SecretSvc.AddCard(context.Background(), masterKey, "visa", "4532015112830366", "JOHN DOE", "12/26", "123", "Tinkoff", ""))

	cmd := NewGetCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"card", "visa", "--master-password", testMasterPwd})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), "JOHN DOE")
	assert.Contains(t, out.String(), "***") // CVV masked by default
}

func TestGetCardCmd_ShowCVV(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)
	masterKey, err := app.AuthSvc.DeriveMasterKey(context.Background(), testMasterPwd)
	require.NoError(t, err)
	require.NoError(t, app.SecretSvc.AddCard(context.Background(), masterKey, "visa", "4532015112830366", "JOHN DOE", "12/26", "123", "", ""))

	cmd := NewGetCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"card", "visa", "--master-password", testMasterPwd, "--show-cvv"})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), "123")
}

func TestGetTextCmd(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)
	masterKey, err := app.AuthSvc.DeriveMasterKey(context.Background(), testMasterPwd)
	require.NoError(t, err)
	require.NoError(t, app.SecretSvc.AddText(context.Background(), masterKey, "note", "secret note content", ""))

	cmd := NewGetCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"text", "note", "--master-password", testMasterPwd})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), "secret note content")
}

func TestGetBinaryCmd(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)
	masterKey, err := app.AuthSvc.DeriveMasterKey(context.Background(), testMasterPwd)
	require.NoError(t, err)
	fileData := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	require.NoError(t, app.SecretSvc.AddBinary(context.Background(), masterKey, "blob", "blob.bin", fileData, ""))

	outPath := filepath.Join(t.TempDir(), "out.bin")

	cmd := NewGetCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"binary", "blob", "--master-password", testMasterPwd, "--output", outPath})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), "Saved to")
	got, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Equal(t, fileData, got)
}

func TestMaskCardNumber(t *testing.T) {
	assert.Equal(t, "**** **** **** 3456", maskCardNumber("1234567890123456", false))
	assert.Equal(t, "1234567890123456", maskCardNumber("1234567890123456", true))
	assert.Equal(t, "123", maskCardNumber("123", false)) // too short, returned as-is
}
