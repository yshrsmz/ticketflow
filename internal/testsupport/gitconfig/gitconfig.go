// Package gitconfig provides utilities for configuring git in test environments.
// It ensures consistent git configuration across all tests, avoiding issues with
// global git configuration that might affect test results.
package gitconfig

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// Executor is the minimal interface needed to run git configuration commands.
type Executor interface {
	Exec(ctx context.Context, args ...string) (string, error)
}

// Options customizes the git configuration applied to a test repository.
type Options struct {
	UserName             string
	UserEmail            string
	DisableSigning       bool
	DefaultBranch        string
	SetInitDefaultBranch bool
}

func defaultOptions() Options {
	return Options{
		UserName:             "Test User",
		UserEmail:            "test@example.com",
		DisableSigning:       true,
		DefaultBranch:        "",
		SetInitDefaultBranch: false,
	}
}

// Apply configures git for tests in a consistent manner using the provided executor.
func Apply(tb testing.TB, exec Executor, opts ...Options) {
	tb.Helper()

	options := defaultOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	ctx := context.Background()
	commands := [][]string{
		{"config", "user.name", options.UserName},
		{"config", "user.email", options.UserEmail},
	}

	if options.DisableSigning {
		commands = append(commands, []string{"config", "commit.gpgSign", "false"})
	}

	if options.SetInitDefaultBranch && options.DefaultBranch != "" {
		commands = append(commands, []string{"config", "init.defaultBranch", options.DefaultBranch})
	}

	for _, args := range commands {
		_, err := exec.Exec(ctx, args...)
		require.NoError(tb, err, "failed to run git %v", args)
	}
}
