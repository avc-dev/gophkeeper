package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("GOPHKEEPER_SERVER", "")
	t.Setenv("GOPHKEEPER_DB", "")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "localhost:8080", cfg.ServerAddr)
	assert.NotEmpty(t, cfg.DBPath) // derived from UserConfigDir
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("GOPHKEEPER_SERVER", "myserver:9090")
	t.Setenv("GOPHKEEPER_DB", "/tmp/test.db")
	t.Setenv("GOPHKEEPER_TLS", "true")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "myserver:9090", cfg.ServerAddr)
	assert.Equal(t, "/tmp/test.db", cfg.DBPath)
	assert.True(t, cfg.TLSEnabled)
}

func TestLoad_DBPath_DefaultsToUserConfigDir(t *testing.T) {
	t.Setenv("GOPHKEEPER_DB", "")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Contains(t, cfg.DBPath, "gophkeeper")
	assert.Contains(t, cfg.DBPath, "local.db")
}
