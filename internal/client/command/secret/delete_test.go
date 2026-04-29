package secret

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteCmd_Success(t *testing.T) {
	app, _, _ := newTestApp(t)
	addTestCredential(t, app, "github")

	// need authed context; loginApp was already called by addTestCredential
	cmd := NewDeleteCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"credential", "github"})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	assert.Contains(t, out.String(), "deleted")
}

func TestDeleteCmd_NotFound(t *testing.T) {
	app, _, _ := newTestApp(t)
	loginApp(t, app)

	cmd := NewDeleteCmd(app)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"credential", "nonexistent"})

	require.Error(t, cmd.ExecuteContext(context.Background()))
}
