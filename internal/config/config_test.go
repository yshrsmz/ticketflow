package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	assert.Equal(t, "main", cfg.Git.DefaultBranch)
	assert.True(t, cfg.Worktree.Enabled)
	assert.Equal(t, "../.worktrees", cfg.Worktree.BaseDir)
	assert.Equal(t, "tickets", cfg.Tickets.Dir)
	assert.Equal(t, "todo", cfg.Tickets.TodoDir)
	assert.Equal(t, "doing", cfg.Tickets.DoingDir)
	assert.Equal(t, "done", cfg.Tickets.DoneDir)
	assert.Equal(t, "text", cfg.Output.DefaultFormat)
	assert.True(t, cfg.Output.JSONPretty)
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr string
	}{
		{
			name:    "valid config",
			config:  *Default(),
			wantErr: "",
		},
		{
			name: "empty git.default_branch",
			config: Config{
				Git:     GitConfig{DefaultBranch: ""},
				Tickets: TicketsConfig{Dir: "tickets"},
				Output:  OutputConfig{DefaultFormat: "text"},
			},
			wantErr: "git.default_branch",
		},
		{
			name: "empty tickets.dir",
			config: Config{
				Git:     GitConfig{DefaultBranch: "main"},
				Tickets: TicketsConfig{Dir: ""},
				Output:  OutputConfig{DefaultFormat: "text"},
			},
			wantErr: "tickets.dir",
		},
		{
			name: "invalid output format",
			config: Config{
				Git:     GitConfig{DefaultBranch: "main"},
				Tickets: TicketsConfig{Dir: "tickets"},
				Output:  OutputConfig{DefaultFormat: "xml"},
			},
			wantErr: "output.default_format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".ticketflow.yaml")

	// Save default config
	cfg := Default()
	err := cfg.Save(configPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Load config
	loaded, err := Load(tmpDir)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, cfg.Git.DefaultBranch, loaded.Git.DefaultBranch)
	assert.Equal(t, cfg.Worktree.Enabled, loaded.Worktree.Enabled)
	assert.Equal(t, cfg.Tickets.Dir, loaded.Tickets.Dir)
	assert.Equal(t, cfg.Output.DefaultFormat, loaded.Output.DefaultFormat)
}

func TestLoadNonExistentConfig(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := Load(tmpDir)
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "configuration file not found")
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".ticketflow.yaml")

	// Write invalid YAML
	err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644)
	require.NoError(t, err)

	_, err = Load(tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestGetPaths(t *testing.T) {
	cfg := Default()
	projectRoot := "/home/user/project"

	// Test relative paths
	assert.Equal(t, "/home/user/project/tickets", cfg.GetTicketsPath(projectRoot))
	assert.Equal(t, "/home/user/project/tickets/todo", cfg.GetTodoPath(projectRoot))
	assert.Equal(t, "/home/user/project/tickets/doing", cfg.GetDoingPath(projectRoot))
	assert.Equal(t, "/home/user/project/tickets/done", cfg.GetDonePath(projectRoot))
	assert.Equal(t, "/home/user/.worktrees", cfg.GetWorktreePath(projectRoot))

	// Test absolute paths
	cfg.Tickets.Dir = "/absolute/tickets"
	cfg.Worktree.BaseDir = "/absolute/worktrees"

	assert.Equal(t, "/absolute/tickets", cfg.GetTicketsPath(projectRoot))
	assert.Equal(t, "/absolute/worktrees", cfg.GetWorktreePath(projectRoot))
}
