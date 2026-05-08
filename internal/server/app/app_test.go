package app

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/avc-dev/gophkeeper/internal/server/config"
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

// ─── Run ─────────────────────────────────────────────────────────────────────

func TestRun_DBUnavailable(t *testing.T) {
	// Port 19999 on localhost is almost certainly not running Postgres.
	// pgxpool.New succeeds (lazy), but Ping fails immediately with connection refused.
	cfg := config.Config{
		DSN: "postgres://localhost:19999/testdb",
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	err := Run(cfg, log)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "run:")
}

// ─── loadEdKeys ───────────────────────────────────────────────────────────────

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

func TestLoadEdKeys_InvalidPrivateDER(t *testing.T) {
	dir := t.TempDir()
	_, pubPath := writeKeyPair(t, dir)

	// correct block type, but DER bytes are garbage → ParsePKCS8PrivateKey fails
	privPath := filepath.Join(dir, "bad_der_private.pem")
	require.NoError(t, os.WriteFile(privPath,
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("not-valid-der")}),
		0o600))

	_, _, err := loadEdKeys(privPath, pubPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse private key")
}

func TestLoadEdKeys_InvalidPublicDER(t *testing.T) {
	dir := t.TempDir()
	privPath, _ := writeKeyPair(t, dir)

	// correct block type, but DER bytes are garbage → ParsePKIXPublicKey fails
	pubPath := filepath.Join(dir, "bad_der_public.pem")
	require.NoError(t, os.WriteFile(pubPath,
		pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: []byte("not-valid-der")}),
		0o600))

	_, _, err := loadEdKeys(privPath, pubPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse public key")
}

func TestLoadEdKeys_WrongPrivateBlockType(t *testing.T) {
	dir := t.TempDir()
	_, pubPath := writeKeyPair(t, dir)

	// Valid PKCS8 DER but mislabelled block type — triggers "not PRIVATE KEY" check
	_, realPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(realPriv)
	require.NoError(t, err)
	privPath := filepath.Join(dir, "wrong_type.pem")
	require.NoError(t, os.WriteFile(privPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600))

	_, _, err = loadEdKeys(privPath, pubPath)
	require.Error(t, err)
}

func TestLoadEdKeys_PrivateKeyNotEd25519(t *testing.T) {
	dir := t.TempDir()
	_, pubPath := writeKeyPair(t, dir)

	// RSA key in valid PKCS8 format — passes ParsePKCS8PrivateKey but fails type assertion
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(rsaKey)
	require.NoError(t, err)
	privPath := filepath.Join(dir, "rsa_private.pem")
	require.NoError(t, os.WriteFile(privPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), 0o600))

	_, _, err = loadEdKeys(privPath, pubPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not Ed25519")
}

func TestLoadEdKeys_WrongPublicBlockType(t *testing.T) {
	dir := t.TempDir()
	privPath, _ := writeKeyPair(t, dir)

	// Valid PKIX DER but mislabelled block type — triggers "not PUBLIC KEY" check
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKIXPublicKey(pub)
	require.NoError(t, err)
	pubPath := filepath.Join(dir, "wrong_pub_type.pem")
	require.NoError(t, os.WriteFile(pubPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600))

	_, _, err = loadEdKeys(privPath, pubPath)
	require.Error(t, err)
}

func TestLoadEdKeys_PublicKeyNotEd25519(t *testing.T) {
	dir := t.TempDir()
	privPath, _ := writeKeyPair(t, dir)

	// RSA public key in valid PKIX format — passes ParsePKIXPublicKey but fails type assertion
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	der, err := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
	require.NoError(t, err)
	pubPath := filepath.Join(dir, "rsa_public.pem")
	require.NoError(t, os.WriteFile(pubPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}), 0o600))

	_, _, err = loadEdKeys(privPath, pubPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not Ed25519")
}
