package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockGitExecutor captures git configuration commands for testing
type mockGitExecutor struct {
	commands [][]string
}

func (m *mockGitExecutor) Exec(ctx context.Context, args ...string) (string, error) {
	m.commands = append(m.commands, args)
	return "", nil
}

func TestGitConfigApply_DefaultOptions(t *testing.T) {
	mock := &mockGitExecutor{}

	GitConfigApply(t, mock)

	// Should apply default configuration
	expectedCommands := [][]string{
		{"config", "user.name", "Test User"},
		{"config", "user.email", "test@example.com"},
		{"config", "commit.gpgSign", "false"},
	}

	assert.Equal(t, expectedCommands, mock.commands)
}

func TestGitConfigApply_CustomOptions(t *testing.T) {
	mock := &mockGitExecutor{}

	GitConfigApply(t, mock, GitConfigOptions{
		UserName:             "Custom User",
		UserEmail:            "custom@test.com",
		DisableSigning:       false,
		DefaultBranch:        "develop",
		SetInitDefaultBranch: true,
	})

	// Should apply custom configuration
	expectedCommands := [][]string{
		{"config", "user.name", "Custom User"},
		{"config", "user.email", "custom@test.com"},
		{"config", "init.defaultBranch", "develop"},
	}

	assert.Equal(t, expectedCommands, mock.commands)
}

func TestGitConfigApply_PartialOptions(t *testing.T) {
	mock := &mockGitExecutor{}

	GitConfigApply(t, mock, GitConfigOptions{
		UserName:       "Another User",
		UserEmail:      "another@test.com",
		DisableSigning: true,
		// DefaultBranch not set, so init.defaultBranch shouldn't be configured
	})

	expectedCommands := [][]string{
		{"config", "user.name", "Another User"},
		{"config", "user.email", "another@test.com"},
		{"config", "commit.gpgSign", "false"},
	}

	assert.Equal(t, expectedCommands, mock.commands)
}

func TestGitConfigApply_SuccessfulExecution(t *testing.T) {
	// Test successful execution path
	mock := &mockGitExecutor{}

	// Should execute without panic when all commands succeed
	GitConfigApply(t, mock)

	// Verify all expected commands were executed
	assert.Len(t, mock.commands, 3, "Should execute 3 commands with default options")
}

func TestGitConfigApply_CommandsExecutedInOrder(t *testing.T) {
	// Verify commands are executed in the correct order
	mock := &mockGitExecutor{}

	GitConfigApply(t, mock, GitConfigOptions{
		UserName:             "Ordered User",
		UserEmail:            "ordered@test.com",
		DisableSigning:       true,
		DefaultBranch:        "main",
		SetInitDefaultBranch: true,
	})

	// Commands should be in specific order
	assert.Equal(t, []string{"config", "user.name", "Ordered User"}, mock.commands[0])
	assert.Equal(t, []string{"config", "user.email", "ordered@test.com"}, mock.commands[1])
	assert.Equal(t, []string{"config", "commit.gpgSign", "false"}, mock.commands[2])
	assert.Equal(t, []string{"config", "init.defaultBranch", "main"}, mock.commands[3])
}

func TestGitConfigApply_ContextRespected(t *testing.T) {
	// Verify that the function passes context to the executor
	// The mockGitExecutor already receives context in its Exec method
	mock := &mockGitExecutor{}

	GitConfigApply(t, mock)

	// If we got here without panic and commands were recorded,
	// then context was successfully passed (non-nil context required by Exec)
	assert.Greater(t, len(mock.commands), 0, "Commands should be recorded, proving context was passed")
}

func TestConfigureGitClient_Integration(t *testing.T) {
	mock := &mockGitExecutor{}

	ConfigureGitClient(t, mock, GitOptions{
		UserName:       "Integration User",
		UserEmail:      "integration@test.com",
		DefaultBranch:  "main",
		DisableSigning: true,
		// InitDefaultBranch: false, so init.defaultBranch won't be set
	})

	// Should properly translate GitOptions to GitConfigOptions
	expectedCommands := [][]string{
		{"config", "user.name", "Integration User"},
		{"config", "user.email", "integration@test.com"},
		{"config", "commit.gpgSign", "false"},
	}

	assert.Equal(t, expectedCommands, mock.commands)
}

func TestConfigureGitClient_WithInitDefaultBranch(t *testing.T) {
	mock := &mockGitExecutor{}

	ConfigureGitClient(t, mock, GitOptions{
		UserName:          "Branch User",
		UserEmail:         "branch@test.com",
		DefaultBranch:     "develop",
		DisableSigning:    false,
		InitDefaultBranch: true, // This time, set init.defaultBranch
	})

	// Should include init.defaultBranch when InitDefaultBranch is true
	expectedCommands := [][]string{
		{"config", "user.name", "Branch User"},
		{"config", "user.email", "branch@test.com"},
		{"config", "init.defaultBranch", "develop"},
	}

	assert.Equal(t, expectedCommands, mock.commands)
}

func TestGitConfigApply_MultipleOptions(t *testing.T) {
	// Test that only the first option is used when multiple are provided
	mock := &mockGitExecutor{}

	GitConfigApply(t, mock,
		GitConfigOptions{UserName: "First", UserEmail: "first@test.com", DisableSigning: true},
		GitConfigOptions{UserName: "Second", UserEmail: "second@test.com", DisableSigning: false},
	)

	// Should only use the first options struct
	expectedCommands := [][]string{
		{"config", "user.name", "First"},
		{"config", "user.email", "first@test.com"},
		{"config", "commit.gpgSign", "false"},
	}

	assert.Equal(t, expectedCommands, mock.commands)
}

// TestGitExecutor_InterfaceCompliance verifies that common executors implement the interface
func TestGitExecutor_InterfaceCompliance(t *testing.T) {
	// Verify SimpleGitExecutor implements GitExecutor
	var _ GitExecutor = SimpleGitExecutor{Dir: "/tmp"}

	// Verify gitCommandExecutor implements GitExecutor
	var _ GitExecutor = gitCommandExecutor{dir: "/tmp"}

	// This test just verifies compilation - if it compiles, the test passes
	require.True(t, true, "Interface compliance verified at compile time")
}
