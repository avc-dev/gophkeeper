package app

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeKeyPair generates an Ed25519 key pair and writes both PEM files to dir.
func writeKeyPair(t *testing.T, dir string) (privPath, pubPath string) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)
	privPath = filepath.Join(dir, "private.pem")
	require.NoError(t, os.WriteFile(privPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}), 0o600))

	pubDER, err := x509.MarshalPKIXPublicKey(pub)
	require.NoError(t, err)
	pubPath = filepath.Join(dir, "public.pem")
	require.NoError(t, os.WriteFile(pubPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}), 0o600))

	return privPath, pubPath
}

func TestLoadEdKeys_Success(t *testing.T) {
	dir := t.TempDir()
	privPath, pubPath := writeKeyPair(t, dir)

	priv, pub, err := loadEdKeys(privPath, pubPath)
	require.NoError(t, err)
	assert.Len(t, priv, ed25519.PrivateKeySize)
	assert.Len(t, pub, ed25519.PublicKeySize)
}

func TestLoadEdKeys_MissingPrivate(t *testing.T) {
	dir := t.TempDir()
	_, pubPath := writeKeyPair(t, dir)

	_, _, err := loadEdKeys(filepath.Join(dir, "nonexistent.pem"), pubPath)
	require.Error(t, err)
}

func TestLoadEdKeys_MissingPublic(t *testing.T) {
	dir := t.TempDir()
	privPath, _ := writeKeyPair(t, dir)

	_, _, err := loadEdKeys(privPath, filepath.Join(dir, "nonexistent.pem"))
	require.Error(t, err)
}

func TestLoadEdKeys_InvalidPrivatePEM(t *testing.T) {
	dir := t.TempDir()
	_, pubPath := writeKeyPair(t, dir)

	privPath := filepath.Join(dir, "bad_private.pem")
	require.NoError(t, os.WriteFile(privPath, []byte("not a pem"), 0o600))

	_, _, err := loadEdKeys(privPath, pubPath)
	require.Error(t, err)
}

func TestLoadEdKeys_InvalidPublicPEM(t *testing.T) {
	dir := t.TempDir()
	privPath, _ := writeKeyPair(t, dir)

	pubPath := filepath.Join(dir, "bad_public.pem")
	require.NoError(t, os.WriteFile(pubPath, []byte("not a pem"), 0o600))

	_, _, err := loadEdKeys(privPath, pubPath)
	require.Error(t, err)
}

func TestLoadEdKeys_WrongKeyType(t *testing.T) {
	// Write private key with wrong PEM block type to trigger "not Ed25519" error
	dir := t.TempDir()
	_, pubPath := writeKeyPair(t, dir)

	// Use PUBLIC KEY block type where PRIVATE KEY is expected
	privPath := filepath.Join(dir, "wrong_type.pem")
	_, realPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(realPriv)
	require.NoError(t, err)
	// deliberately mislabel as "CERTIFICATE" so pem.Decode gives wrong block.Type
	require.NoError(t, os.WriteFile(privPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600))

	_, _, err = loadEdKeys(privPath, pubPath)
	require.Error(t, err)
}
