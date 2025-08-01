package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Git      GitConfig      `yaml:"git"`
	Worktree WorktreeConfig `yaml:"worktree"`
	Tickets  TicketsConfig  `yaml:"tickets"`
	Output   OutputConfig   `yaml:"output"`
	Timeouts TimeoutsConfig `yaml:"timeouts"`
}

// GitConfig represents git-related configuration
type GitConfig struct {
	DefaultBranch string `yaml:"default_branch"`
}

// WorktreeConfig represents worktree-related configuration
type WorktreeConfig struct {
	Enabled      bool     `yaml:"enabled"`
	BaseDir      string   `yaml:"base_dir"`
	InitCommands []string `yaml:"init_commands"`
}

// TicketsConfig represents ticket-related configuration
type TicketsConfig struct {
	Dir      string `yaml:"dir"`
	TodoDir  string `yaml:"todo_dir"`
	DoingDir string `yaml:"doing_dir"`
	DoneDir  string `yaml:"done_dir"`
	Template string `yaml:"template"`
}

// OutputConfig represents output formatting configuration
type OutputConfig struct {
	DefaultFormat string `yaml:"default_format"`
	JSONPretty    bool   `yaml:"json_pretty"`
}

// TimeoutsConfig represents timeout configuration for various operations
type TimeoutsConfig struct {
	Git          int `yaml:"git"`           // Timeout for git operations in seconds
	InitCommands int `yaml:"init_commands"` // Timeout for worktree init commands in seconds
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		Git: GitConfig{
			DefaultBranch: DefaultBranch,
		},
		Worktree: WorktreeConfig{
			Enabled: true,
			BaseDir: DefaultWorktreeBase,
			InitCommands: []string{
				"git fetch origin",
			},
		},
		Tickets: TicketsConfig{
			Dir:      DefaultTicketsDir,
			TodoDir:  DefaultTodoDir,
			DoingDir: DefaultDoingDir,
			DoneDir:  DefaultDoneDir,
			Template: `# Summary

[Describe the ticket summary here]

## Tasks
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

## Technical Specifications

[Add technical details as needed]

## Notes

[Additional notes or remarks]`,
		},
		Output: OutputConfig{
			DefaultFormat: DefaultOutputFormat,
			JSONPretty:    true,
		},
		Timeouts: TimeoutsConfig{
			Git:          DefaultGitTimeoutSeconds,
			InitCommands: DefaultInitCommandsTimeoutSeconds,
		},
	}
}

// Load loads configuration from the specified project root
func Load(projectRoot string) (*Config, error) {
	configPath := filepath.Join(projectRoot, ConfigFileName)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, ticketerrors.ErrConfigNotFound
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, ticketerrors.NewConfigError("file", configPath, fmt.Errorf("failed to read config file: %w", err))
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, ticketerrors.NewConfigError("format", "yaml", fmt.Errorf("failed to parse config file: %w", err))
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, ticketerrors.NewConfigError("validation", "", fmt.Errorf("invalid configuration: %w", err))
	}

	return &config, nil
}

// Save saves the configuration to the specified path
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, DirPermission); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, FilePermission); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate Git config
	if c.Git.DefaultBranch == "" {
		return ticketerrors.NewConfigError("git.default_branch", "", ticketerrors.ErrConfigInvalid)
	}

	// Validate Tickets config
	if c.Tickets.Dir == "" {
		return ticketerrors.NewConfigError("tickets.dir", "", ticketerrors.ErrConfigInvalid)
	}

	// Validate Output config
	if c.Output.DefaultFormat != FormatText && c.Output.DefaultFormat != FormatJSON {
		return ticketerrors.NewConfigError("output.default_format", c.Output.DefaultFormat, ticketerrors.ErrConfigInvalid)
	}

	// Validate Timeouts config
	if err := validateTimeout(c.Timeouts.Git, "timeouts.git"); err != nil {
		return err
	}
	if err := validateTimeout(c.Timeouts.InitCommands, "timeouts.init_commands"); err != nil {
		return err
	}

	return nil
}

// GetTicketsPath returns the full path to the tickets directory
func (c *Config) GetTicketsPath(projectRoot string) string {
	if filepath.IsAbs(c.Tickets.Dir) {
		return c.Tickets.Dir
	}
	return filepath.Join(projectRoot, c.Tickets.Dir)
}

// GetTodoPath returns the full path to the todo directory
func (c *Config) GetTodoPath(projectRoot string) string {
	return filepath.Join(c.GetTicketsPath(projectRoot), c.Tickets.TodoDir)
}

// GetDoingPath returns the full path to the doing directory
func (c *Config) GetDoingPath(projectRoot string) string {
	return filepath.Join(c.GetTicketsPath(projectRoot), c.Tickets.DoingDir)
}

// GetDonePath returns the full path to the done directory
func (c *Config) GetDonePath(projectRoot string) string {
	return filepath.Join(c.GetTicketsPath(projectRoot), c.Tickets.DoneDir)
}

// GetWorktreePath returns the full path to the worktree base directory
func (c *Config) GetWorktreePath(projectRoot string) string {
	if filepath.IsAbs(c.Worktree.BaseDir) {
		return c.Worktree.BaseDir
	}
	return filepath.Join(projectRoot, c.Worktree.BaseDir)
}

// GetGitTimeout returns the timeout duration for git operations
func (c *Config) GetGitTimeout() time.Duration {
	if c.Timeouts.Git <= 0 {
		return DefaultGitTimeout
	}
	return time.Duration(c.Timeouts.Git) * time.Second
}

// GetInitCommandsTimeout returns the timeout duration for init commands
func (c *Config) GetInitCommandsTimeout() time.Duration {
	if c.Timeouts.InitCommands <= 0 {
		return DefaultInitCommandsTimeout
	}
	return time.Duration(c.Timeouts.InitCommands) * time.Second
}

// validateTimeout validates a timeout value is within acceptable range
func validateTimeout(value int, fieldName string) error {
	if value < 0 {
		return ticketerrors.NewConfigError(fieldName, fmt.Sprintf("%d", value), ticketerrors.ErrConfigInvalid)
	}
	if value > MaxTimeoutSeconds {
		return ticketerrors.NewConfigError(fieldName,
			fmt.Sprintf("%d exceeds maximum of %d seconds", value, MaxTimeoutSeconds),
			ticketerrors.ErrConfigInvalid)
	}
	return nil
}
