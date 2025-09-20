package testutil

import (
	"bytes"
	"context"
	"os/exec"
)

// SimpleGitExecutor provides a simple implementation of gitconfig.Executor
// for tests that need to configure git in a specific directory.
type SimpleGitExecutor struct {
	Dir string
}

// Exec runs a git command in the configured directory.
func (e SimpleGitExecutor) Exec(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = e.Dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	return out.String(), cmd.Run()
}
