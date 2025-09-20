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
// The implementation combines stdout and stderr into a single output buffer.
// This simplifies test usage but means error messages are mixed with regular output.
// The combined output is returned along with any execution error.
func (e SimpleGitExecutor) Exec(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = e.Dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	return out.String(), cmd.Run()
}
