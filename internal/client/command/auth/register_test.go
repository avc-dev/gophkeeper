package auth

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterCmd_Success(t *testing.T) {
	authMock := &mockAuthGRPC{} // registerErr == nil → success
	app := newTestApp(t, authMock, &mockSecretsGRPC{})

	cmd := NewRegisterCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--email", "new@example.com", "--password", "secret"})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), "Registration successful")
}

func TestRegisterCmd_ServerError(t *testing.T) {
	authMock := &mockAuthGRPC{registerErr: errors.New("email taken")}
	app := newTestApp(t, authMock, &mockSecretsGRPC{})

	cmd := NewRegisterCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--email", "taken@example.com", "--password", "pass"})

	err := cmd.ExecuteContext(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "registration failed")
}

func TestRegisterCmd_MissingEmail(t *testing.T) {
	app := newTestApp(t, &mockAuthGRPC{}, &mockSecretsGRPC{})

	cmd := NewRegisterCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--password", "pass"})

	require.Error(t, cmd.ExecuteContext(context.Background()))
}
