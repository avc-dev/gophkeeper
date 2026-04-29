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

func TestAddCredentialCmd(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)

	cmd := NewAddCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"credential",
		"--name", "github",
		"--login", "alice",
		"--password", "s3cr3t",
		"--master-password", testMasterPwd,
	})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), `"github" saved`)
}

func TestAddCredentialCmd_MissingRequired(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)

	cmd := NewAddCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"credential", "--name", "gh", "--master-password", testMasterPwd})

	require.Error(t, cmd.ExecuteContext(context.Background()))
}

func TestAddCardCmd(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)

	cmd := NewAddCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"card",
		"--name", "visa",
		"--number", "4532015112830366",
		"--holder", "JOHN DOE",
		"--expiry", "12/26",
		"--cvv", "123",
		"--master-password", testMasterPwd,
	})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), `"visa" saved`)
}

func TestAddCardCmd_InvalidCard(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)

	cmd := NewAddCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"card",
		"--name", "bad",
		"--number", "4532015112830367", // fails Luhn (valid length, wrong checksum)
		"--holder", "JOHN",
		"--expiry", "12/26",
		"--cvv", "123",
		"--master-password", testMasterPwd,
	})

	require.Error(t, cmd.ExecuteContext(context.Background()))
}

func TestAddTextCmd(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)

	cmd := NewAddCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"text",
		"--name", "note",
		"--content", "my secret text",
		"--master-password", testMasterPwd,
	})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), `"note" saved`)
}

func TestAddTextCmd_FromFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "note.txt")
	require.NoError(t, os.WriteFile(tmp, []byte("file content"), 0o600))

	app, _, _ := newTestApp(t)
	loginApp(t, app)

	cmd := NewAddCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"text",
		"--name", "note2",
		"--file", tmp,
		"--master-password", testMasterPwd,
	})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), `"note2" saved`)
}

func TestAddTextCmd_MissingContent(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)

	cmd := NewAddCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"text", "--name", "note", "--master-password", testMasterPwd})

	require.Error(t, cmd.ExecuteContext(context.Background()))
}

func TestAddBinaryCmd(t *testing.T) {
	// create a temp file
	tmp := filepath.Join(t.TempDir(), "data.bin")
	require.NoError(t, os.WriteFile(tmp, []byte{0x01, 0x02, 0x03}, 0o600))

	app, _, _ := newTestApp(t)
	loginApp(t, app)

	cmd := NewAddCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"binary",
		"--name", "myfile",
		"--file", tmp,
		"--master-password", testMasterPwd,
	})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), `"myfile" saved`)
}
