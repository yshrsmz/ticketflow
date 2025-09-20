package testutil_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/testutil"
)

func TestGitConfigApply_DefaultOptions_Integration(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	g := git.New(tmpDir)
	ctx := context.Background()

	_, err := g.Exec(ctx, "init")
	require.NoError(t, err)

	testutil.GitConfigApply(t, g)

	name, err := g.Exec(ctx, "config", "--local", "--get", "user.name")
	require.NoError(t, err)
	assert.Equal(t, "Test User", strings.TrimSpace(name))

	email, err := g.Exec(ctx, "config", "--local", "--get", "user.email")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", strings.TrimSpace(email))

	signing, err := g.Exec(ctx, "config", "--local", "--get", "commit.gpgSign")
	require.NoError(t, err)
	assert.Equal(t, "false", strings.TrimSpace(signing))
}

func TestGitConfigApply_CustomOptions_Integration(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	g := git.New(tmpDir)
	ctx := context.Background()

	_, err := g.Exec(ctx, "init")
	require.NoError(t, err)

	testutil.GitConfigApply(t, g, testutil.GitConfigOptions{
		UserName:             "Alice",
		UserEmail:            "alice@example.com",
		DisableSigning:       false,
		DefaultBranch:        "develop",
		SetInitDefaultBranch: true,
	})

	name, err := g.Exec(ctx, "config", "--local", "--get", "user.name")
	require.NoError(t, err)
	assert.Equal(t, "Alice", strings.TrimSpace(name))

	email, err := g.Exec(ctx, "config", "--local", "--get", "user.email")
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", strings.TrimSpace(email))

	_, err = g.Exec(ctx, "config", "--local", "--get", "commit.gpgSign")
	assert.Error(t, err)

	initBranch, err := g.Exec(ctx, "config", "--local", "--get", "init.defaultBranch")
	require.NoError(t, err)
	assert.Equal(t, "develop", strings.TrimSpace(initBranch))
}
