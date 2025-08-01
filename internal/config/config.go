package config

import (
	"fmt"
	"os"
	"path/filepath"

	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Git      GitConfig      `yaml:"git"`
	Worktree WorktreeConfig `yaml:"worktree"`
	Tickets  TicketsConfig  `yaml:"tickets"`
	Output   OutputConfig   `yaml:"output"`
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

// Default returns the default configuration
func Default() *Config {
	return &Config{
		Git: GitConfig{
			DefaultBranch: "main",
		},
		Worktree: WorktreeConfig{
			Enabled: true,
			BaseDir: "../.worktrees",
			InitCommands: []string{
				"git fetch origin",
			},
		},
		Tickets: TicketsConfig{
			Dir:      "tickets",
			TodoDir:  "todo",
			DoingDir: "doing",
			DoneDir:  "done",
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
			DefaultFormat: "text",
			JSONPretty:    true,
		},
	}
}

// Load loads configuration from the specified project root
func Load(projectRoot string) (*Config, error) {
	configPath := filepath.Join(projectRoot, ".ticketflow.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, ticketerrors.ErrConfigNotFound
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, ticketerrors.NewConfigError("", "", fmt.Errorf("failed to read config file: %w", err))
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, ticketerrors.NewConfigError("", "", fmt.Errorf("failed to parse config file: %w", err))
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, ticketerrors.NewConfigError("", "", fmt.Errorf("invalid configuration: %w", err))
	}

	return &config, nil
}

// Save saves the configuration to the specified path
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
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
	if c.Output.DefaultFormat != "text" && c.Output.DefaultFormat != "json" {
		return ticketerrors.NewConfigError("output.default_format", c.Output.DefaultFormat, ticketerrors.ErrConfigInvalid)
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
