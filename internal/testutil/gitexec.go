package testutil

import (
	"bytes"
	"context"
	"os/exec"
)

// SimpleGitExecutor provides a simple implementation of gitconfig.Executor
// for tests that need to configure git in a specific directory.
// Note: This implementation combines stdout and stderr into a single output
// buffer for simplicity in test scenarios. If you need to distinguish between
// stdout and stderr, consider using a more sophisticated executor.
type SimpleGitExecutor struct {
	Dir string
}

// Exec runs a git command in the configured directory.
// The implementation intentionally combines stdout and stderr into a single output buffer
// to match the gitconfig.Executor interface signature which requires (string, error).
// This simplifies test usage where exact output separation is not critical.
// For tests requiring separate stdout/stderr, use GitRepo methods instead.
func (e SimpleGitExecutor) Exec(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = e.Dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	return out.String(), cmd.Run()
}
