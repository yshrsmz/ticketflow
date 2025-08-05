package testutil

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
)

// SetupTicketflowProject creates a complete ticketflow project structure
func SetupTicketflowProject(t *testing.T, dir string, opts ...ProjectOption) {
	t.Helper()

	// Default options
	options := projectOptions{
		createConfig: true,
		createDirs:   true,
		gitInit:      true,
	}

	for _, opt := range opts {
		opt(&options)
	}

	// Create directory structure
	if options.createDirs {
		CreateTicketDirs(t, dir)
	}

	// Create config file
	if options.createConfig {
		CreateConfigFile(t, dir, options.config)
	}

	// Initialize git if requested
	if options.gitInit {
		SetupGitRepo(t, dir)
	}
}

// CreateTicketDirs creates the standard ticket directory structure
func CreateTicketDirs(t *testing.T, baseDir string) {
	t.Helper()

	dirs := []string{
		"tickets/todo",
		"tickets/doing",
		"tickets/done",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(baseDir, dir), 0755)
		require.NoError(t, err, "Failed to create directory: %s", dir)
	}
}

// CreateConfigFile creates a .ticketflow.yaml config file with default configuration
// TODO: In the future, use the cfg parameter to generate custom config content
func CreateConfigFile(t *testing.T, dir string, cfg *config.Config) {
	t.Helper()

	// Currently using hardcoded config, cfg parameter reserved for future use

	configPath := filepath.Join(dir, ".ticketflow.yaml")
	content := `git:
  default_branch: main

worktree:
  enabled: true
  base_dir: "../test.worktrees"
  init_commands:
    - git fetch origin

tickets:
  dir: tickets
  todo_dir: tickets/todo
  doing_dir: tickets/doing
  done_dir: tickets/done
  template: |
    # Summary
    
    [Describe the ticket summary here]
    
    ## Tasks
    - [ ] Task 1
    - [ ] Task 2
    - [ ] Task 3

output:
  default_format: text
  json_pretty: true

timeouts:
  git: 300
  init_commands: 600
`

	err := os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err, "Failed to create config file")
}

// CreateTicketFile creates a ticket file with frontmatter
func CreateTicketFile(t *testing.T, dir string, ticketID string, status string, opts ...TicketFileOption) string {
	t.Helper()

	// Determine directory based on status
	statusDir := "todo"
	switch status {
	case "doing":
		statusDir = "doing"
	case "done":
		statusDir = "done"
	}

	// Create ticket with options
	ticketOpts := []TicketOption{WithID(ticketID)}
	for _, opt := range opts {
		ticketOpts = append(ticketOpts, opt.ToTicketOptions()...)
	}
	ticket := TicketFixture(ticketOpts...)

	// Generate content
	content := TicketContent(ticket.Priority, ticket.Description, ticket.CreatedAt.Time, map[string]interface{}{
		"started_at": ticket.StartedAt.Time,
		"closed_at":  ticket.ClosedAt.Time,
		"related":    ticket.Related,
	})

	// Write file
	ticketPath := filepath.Join(dir, "tickets", statusDir, ticketID+".md")
	err := os.WriteFile(ticketPath, []byte(content), 0644)
	require.NoError(t, err, "Failed to create ticket file")

	return ticketPath
}

// AssertFileExists asserts that a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	require.NoError(t, err, "File should exist: %s", path)
}

// AssertFileNotExists asserts that a file does not exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	require.Error(t, err, "File should not exist: %s", path)
	require.True(t, os.IsNotExist(err), "Error should be 'file not found'")
}

// AssertDirExists asserts that a directory exists
func AssertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err, "Directory should exist: %s", path)
	require.True(t, info.IsDir(), "Path should be a directory: %s", path)
}

// AssertSymlinkTarget asserts that a symlink points to the expected target
func AssertSymlinkTarget(t *testing.T, linkPath, expectedTarget string) {
	t.Helper()

	target, err := os.Readlink(linkPath)
	require.NoError(t, err, "Failed to read symlink: %s", linkPath)
	require.Equal(t, expectedTarget, target, "Symlink target mismatch")
}

// ReadFileContent reads and returns file content
func ReadFileContent(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err, "Failed to read file: %s", path)
	return string(content)
}

// WriteFileContent writes content to a file
func WriteFileContent(t *testing.T, path, content string) {
	t.Helper()
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err, "Failed to write file: %s", path)
}

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "ticketflow-test-*")
	require.NoError(t, err, "Failed to create temp directory")

	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	return tmpDir
}

// ChDir changes to directory and restores original on cleanup
func ChDir(t *testing.T, dir string) {
	t.Helper()

	originalWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current directory")

	err = os.Chdir(dir)
	require.NoError(t, err, "Failed to change directory")

	t.Cleanup(func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err, "Failed to restore original directory")
	})
}

// projectOptions holds options for project setup
type projectOptions struct {
	createConfig bool
	createDirs   bool
	gitInit      bool
	config       *config.Config
}

// ProjectOption customizes project setup
type ProjectOption func(*projectOptions)

// WithoutConfig skips config file creation
func WithoutConfig() ProjectOption {
	return func(o *projectOptions) {
		o.createConfig = false
	}
}

// WithoutDirs skips directory creation
func WithoutDirs() ProjectOption {
	return func(o *projectOptions) {
		o.createDirs = false
	}
}

// WithoutGit skips git initialization
func WithoutGit() ProjectOption {
	return func(o *projectOptions) {
		o.gitInit = false
	}
}

// WithCustomConfig uses a custom config
func WithCustomConfig(cfg *config.Config) ProjectOption {
	return func(o *projectOptions) {
		o.config = cfg
	}
}

// TicketFileOption represents options for ticket file creation
type TicketFileOption interface {
	ToTicketOptions() []TicketOption
}

// ticketFileOptions implements TicketFileOption
type ticketFileOptions struct {
	opts []TicketOption
}

func (t ticketFileOptions) ToTicketOptions() []TicketOption {
	return t.opts
}

// WithTicketOptions wraps ticket options for file creation
func WithTicketOptions(opts ...TicketOption) TicketFileOption {
	return ticketFileOptions{opts: opts}
}

// CountFiles counts files matching a pattern
func CountFiles(t *testing.T, dir string, pattern string) int {
	t.Helper()

	var count int
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if matched {
				count++
			}
		}
		return nil
	})
	require.NoError(t, err, "Failed to walk directory")

	return count
}
