package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("JWT_PRIVATE_KEY_FILE", "/tmp/priv.pem")
	t.Setenv("JWT_PUBLIC_KEY_FILE", "/tmp/pub.pem")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, ":8080", cfg.Addr)
	assert.Equal(t, "/tmp/priv.pem", cfg.JWTPrivateKeyFile)
}

func TestLoad_MissingPrivateKey(t *testing.T) {
	t.Setenv("JWT_PRIVATE_KEY_FILE", "")
	t.Setenv("JWT_PUBLIC_KEY_FILE", "/tmp/pub.pem")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_PRIVATE_KEY_FILE")
}

func TestLoad_MissingPublicKey(t *testing.T) {
	t.Setenv("JWT_PRIVATE_KEY_FILE", "/tmp/priv.pem")
	t.Setenv("JWT_PUBLIC_KEY_FILE", "")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_PUBLIC_KEY_FILE")
}

func TestLoad_TLSPartialConfig(t *testing.T) {
	t.Setenv("JWT_PRIVATE_KEY_FILE", "/tmp/priv.pem")
	t.Setenv("JWT_PUBLIC_KEY_FILE", "/tmp/pub.pem")
	t.Setenv("TLS_CERT_FILE", "/tmp/cert.pem")
	t.Setenv("TLS_KEY_FILE", "")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TLS_CERT_FILE")
}

func TestLoad_TLSBothSet(t *testing.T) {
	t.Setenv("JWT_PRIVATE_KEY_FILE", "/tmp/priv.pem")
	t.Setenv("JWT_PUBLIC_KEY_FILE", "/tmp/pub.pem")
	t.Setenv("TLS_CERT_FILE", "/tmp/cert.pem")
	t.Setenv("TLS_KEY_FILE", "/tmp/key.pem")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "/tmp/cert.pem", cfg.TLSCertFile)
}
