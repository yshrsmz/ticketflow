package gitconfig_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/testsupport/gitconfig"
)

func TestApplyDefaultOptions(t *testing.T) {
	tmpDir := t.TempDir()
	g := git.New(tmpDir)
	ctx := context.Background()

	_, err := g.Exec(ctx, "init")
	require.NoError(t, err)

	gitconfig.Apply(t, g)

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

func TestApplyCustomOptions(t *testing.T) {
	tmpDir := t.TempDir()
	g := git.New(tmpDir)
	ctx := context.Background()

	_, err := g.Exec(ctx, "init")
	require.NoError(t, err)

	gitconfig.Apply(t, g, gitconfig.Options{
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
