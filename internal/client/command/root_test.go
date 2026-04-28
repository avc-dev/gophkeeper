package command

import (
	"bytes"
	"context"
	"testing"

	"github.com/avc-dev/gophkeeper/internal/client/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTransportCredentials_Insecure(t *testing.T) {
	cfg := config.Config{TLSEnabled: false}
	creds, err := buildTransportCredentials(cfg)
	require.NoError(t, err)
	assert.NotNil(t, creds)
	assert.Equal(t, "insecure", creds.Info().SecurityProtocol)
}

func TestBuildTransportCredentials_TLS_NoCA(t *testing.T) {
	cfg := config.Config{TLSEnabled: true}
	creds, err := buildTransportCredentials(cfg)
	require.NoError(t, err)
	assert.Equal(t, "tls", creds.Info().SecurityProtocol)
}

func TestBuildTransportCredentials_TLS_SkipVerify(t *testing.T) {
	cfg := config.Config{TLSEnabled: true, TLSSkipVerify: true}
	creds, err := buildTransportCredentials(cfg)
	require.NoError(t, err)
	assert.Equal(t, "tls", creds.Info().SecurityProtocol)
}

func TestBuildTransportCredentials_TLS_InvalidCAPath(t *testing.T) {
	cfg := config.Config{TLSEnabled: true, TLSCACert: "/nonexistent/ca.pem"}
	_, err := buildTransportCredentials(cfg)
	require.Error(t, err)
}

func TestVersionCmd(t *testing.T) {
	cmd := newVersionCmd("1.2.3", "2024-01-01")
	var out bytes.Buffer
	cmd.SetOut(&out)
	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), "1.2.3")
	assert.Contains(t, out.String(), "2024-01-01")
}
