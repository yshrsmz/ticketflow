package commands_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/testutil"
)

// findProjectRoot traverses up from the current directory to find the project root
// by looking for the go.mod file. This is used by integration tests that need to
// build the ticketflow binary from source.
func findProjectRoot(t *testing.T, startDir string) string {
	projectRoot := startDir
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			return projectRoot
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatal("Could not find project root (go.mod)")
		}
		projectRoot = parent
	}
}

// TestWorkflowCommand_Integration tests the workflow command in a real environment
func TestWorkflowCommand_Integration(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Save current working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err)
	}()

	// Change to temp directory
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	// Configure git locally (not globally) for the test repo
	testutil.GitConfigApply(t, testutil.SimpleGitExecutor{Dir: tmpDir})

	// Build the ticketflow binary
	projectRoot := findProjectRoot(t, originalWd)

	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "ticketflow"), "./cmd/ticketflow")
	buildCmd.Dir = projectRoot
	var buildStderr bytes.Buffer
	buildCmd.Stderr = &buildStderr
	err = buildCmd.Run()
	require.NoError(t, err, "Failed to build ticketflow: %s", buildStderr.String())

	// Run ticketflow workflow command
	workflowCmd := exec.CommandContext(context.Background(), filepath.Join(tmpDir, "ticketflow"), "workflow")
	workflowCmd.Dir = tmpDir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	workflowCmd.Stdout = &stdout
	workflowCmd.Stderr = &stderr

	err = workflowCmd.Run()
	assert.NoError(t, err, "Failed to run workflow command: %s", stderr.String())

	// Verify the output contains expected content
	output := stdout.String()

	// Check for main sections
	assert.Contains(t, output, "# TicketFlow Workflow Guide")
	assert.Contains(t, output, "## Overview")
	assert.Contains(t, output, "## Workflow for Managing Tasks")

	// Check for key workflow steps
	assert.Contains(t, output, "### 1. Create a Feature Ticket")
	assert.Contains(t, output, "ticketflow new my-feature")
	assert.Contains(t, output, "### 2. Start Work on the Ticket")
	assert.Contains(t, output, "ticketflow start <ticket-id>")
	assert.Contains(t, output, "### 3. Navigate to the Worktree")
	assert.Contains(t, output, "cd ../ticketflow.worktrees/<ticket-id>")

	// Check for important commands
	assert.Contains(t, output, "ticketflow close")
	assert.Contains(t, output, "ticketflow cleanup")
	assert.Contains(t, output, "<your-test-command>")
	assert.Contains(t, output, "<your-lint-command>")
	assert.Contains(t, output, "<your-format-command>")

	// Check for integration instructions
	assert.Contains(t, output, "## Integration with Development Tools")
	assert.Contains(t, output, "ticketflow workflow > CLAUDE.md")
	assert.Contains(t, output, "ticketflow workflow >> .cursorrules")

	// Check that output is valid markdown
	assert.True(t, strings.HasPrefix(output, "#"), "Output should start with a markdown header")
	assert.Contains(t, output, "```bash", "Output should contain code blocks")
	assert.Contains(t, output, "```", "Output should have properly closed code blocks")

	// Ensure no error output
	assert.Empty(t, stderr.String(), "Should not produce any error output")
}

// TestWorkflowCommand_OutputRedirection tests that the workflow command output can be redirected
func TestWorkflowCommand_OutputRedirection(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Save current working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err)
	}()

	// Change to temp directory
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	// Configure git locally for the test repo
	testutil.GitConfigApply(t, testutil.SimpleGitExecutor{Dir: tmpDir})

	// Build the ticketflow binary
	projectRoot := findProjectRoot(t, originalWd)

	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "ticketflow"), "./cmd/ticketflow")
	buildCmd.Dir = projectRoot
	err = buildCmd.Run()
	require.NoError(t, err)

	// Test redirecting output to a file
	outputFile := filepath.Join(tmpDir, "CLAUDE.md")
	shellCmd := filepath.Join(tmpDir, "ticketflow") + " workflow > " + outputFile
	redirectCmd := exec.Command("sh", "-c", shellCmd)
	redirectCmd.Dir = tmpDir
	err = redirectCmd.Run()
	require.NoError(t, err)

	// Verify the file was created and contains the expected content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.Contains(t, string(content), "# TicketFlow Workflow Guide")
	assert.Contains(t, string(content), "ticketflow new my-feature")
}
