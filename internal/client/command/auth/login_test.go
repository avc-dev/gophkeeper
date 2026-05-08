package auth

import (
	"bytes"
	"context"
	"errors"
	"testing"

	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginCmd_Success(t *testing.T) {
	authMock := &mockAuthGRPC{
		loginResp: &pb.LoginResponse{Token: "jwt", KdfSalt: make([]byte, 32)},
	}
	// sync после логина упадёт с ошибкой (offline) — это warning, не fatal
	secretsMock := &mockSecretsGRPC{listErr: errors.New("offline")}
	app := newTestApp(t, authMock, secretsMock)

	cmd := NewLoginCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--email", "user@example.com", "--password", "pass"})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), "Logged in successfully")
}

func TestLoginCmd_ServerError(t *testing.T) {
	authMock := &mockAuthGRPC{loginErr: errors.New("unauthorized")}
	app := newTestApp(t, authMock, &mockSecretsGRPC{})

	cmd := NewLoginCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--email", "user@example.com", "--password", "wrong"})

	err := cmd.ExecuteContext(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "login failed")
}

func TestLoginCmd_MissingEmail(t *testing.T) {
	app := newTestApp(t, &mockAuthGRPC{}, &mockSecretsGRPC{})

	cmd := NewLoginCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--password", "pass"})

	err := cmd.ExecuteContext(context.Background())
	require.Error(t, err)
}
