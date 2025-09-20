package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func configureTestGitClient(tb testing.TB, g *Git) {
	tb.Helper()

	steps := [][]string{
		{"config", "user.name", "Test User"},
		{"config", "user.email", "test@example.com"},
		{"config", "commit.gpgSign", "false"},
	}

	ctx := context.Background()
	for _, args := range steps {
		_, err := g.Exec(ctx, args...)
		require.NoError(tb, err)
	}
}
