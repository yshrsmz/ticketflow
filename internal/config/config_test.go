package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	assert.Equal(t, 30, cfg.Timeouts.Git)
	assert.Equal(t, 60, cfg.Timeouts.InitCommands)
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
		{
			name: "negative git timeout",
			config: Config{
				Git:      GitConfig{DefaultBranch: "main"},
				Tickets:  TicketsConfig{Dir: "tickets"},
				Output:   OutputConfig{DefaultFormat: "text"},
				Timeouts: TimeoutsConfig{Git: -1, InitCommands: 60},
			},
			wantErr: "timeouts.git",
		},
		{
			name: "negative init commands timeout",
			config: Config{
				Git:      GitConfig{DefaultBranch: "main"},
				Tickets:  TicketsConfig{Dir: "tickets"},
				Output:   OutputConfig{DefaultFormat: "text"},
				Timeouts: TimeoutsConfig{Git: 30, InitCommands: -1},
			},
			wantErr: "timeouts.init_commands",
		},
		{
			name: "git timeout exceeds maximum",
			config: Config{
				Git:      GitConfig{DefaultBranch: "main"},
				Tickets:  TicketsConfig{Dir: "tickets"},
				Output:   OutputConfig{DefaultFormat: "text"},
				Timeouts: TimeoutsConfig{Git: 3601, InitCommands: 60},
			},
			wantErr: "timeouts.git",
		},
		{
			name: "init commands timeout exceeds maximum",
			config: Config{
				Git:      GitConfig{DefaultBranch: "main"},
				Tickets:  TicketsConfig{Dir: "tickets"},
				Output:   OutputConfig{DefaultFormat: "text"},
				Timeouts: TimeoutsConfig{Git: 30, InitCommands: 3601},
			},
			wantErr: "timeouts.init_commands",
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

func TestLoadWithTimeouts(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".ticketflow.yaml")

	// Create config with custom timeouts
	configYAML := `
git:
  default_branch: main
worktree:
  enabled: true
  base_dir: ../.worktrees
tickets:
  dir: tickets
  todo_dir: todo
  doing_dir: doing
  done_dir: done
output:
  default_format: text
timeouts:
  git: 45
  init_commands: 120
`
	err := os.WriteFile(configPath, []byte(configYAML), 0644)
	require.NoError(t, err)

	// Load and verify
	loaded, err := Load(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, 45, loaded.Timeouts.Git)
	assert.Equal(t, 120, loaded.Timeouts.InitCommands)
	assert.Equal(t, 45*time.Second, loaded.GetGitTimeout())
	assert.Equal(t, 120*time.Second, loaded.GetInitCommandsTimeout())
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

func TestGetTimeouts(t *testing.T) {
	tests := []struct {
		name               string
		gitTimeout         int
		initCommandTimeout int
		wantGit            string
		wantInit           string
	}{
		{
			name:               "default timeouts",
			gitTimeout:         30,
			initCommandTimeout: 60,
			wantGit:            "30s",
			wantInit:           "1m0s",
		},
		{
			name:               "zero timeouts use defaults",
			gitTimeout:         0,
			initCommandTimeout: 0,
			wantGit:            "30s",
			wantInit:           "1m0s",
		},
		{
			name:               "negative timeouts use defaults",
			gitTimeout:         -1,
			initCommandTimeout: -5,
			wantGit:            "30s",
			wantInit:           "1m0s",
		},
		{
			name:               "custom timeouts",
			gitTimeout:         120,
			initCommandTimeout: 300,
			wantGit:            "2m0s",
			wantInit:           "5m0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Timeouts: TimeoutsConfig{
					Git:          tt.gitTimeout,
					InitCommands: tt.initCommandTimeout,
				},
			}

			assert.Equal(t, tt.wantGit, cfg.GetGitTimeout().String())
			assert.Equal(t, tt.wantInit, cfg.GetInitCommandsTimeout().String())
		})
	}
}
