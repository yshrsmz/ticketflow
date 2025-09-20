package cli

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/mocks"
	"github.com/yshrsmz/ticketflow/internal/testutil"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// IMPORTANT: When writing tests that interact with git:
// - Always use local configuration (i.e., run 'git config' without --global; setting cmd.Dir to the test directory ensures this applies to the local repo)
// - Set cmd.Dir to the test directory before running git commands
// - Never modify the user's global git configuration

// Test constants
const (
	testTicketID                  = testutil.TestTicketID
	testDefaultBranch             = testutil.TestDefaultBranch
	expectedOrphanedWorktreeCount = 1
	expectedStaleBranchCount      = 2
)

// testTime parses a time string and fails the test if parsing fails
func testTime(t *testing.T, timeStr string) time.Time {
	parsed, err := time.Parse(time.RFC3339, timeStr)
	require.NoError(t, err)
	return parsed
}

// createDoneTicket creates a ticket with closed status
func createDoneTicket(id string, closedAt time.Time) ticket.Ticket {
	return ticket.Ticket{
		ID:       id,
		ClosedAt: ticket.RFC3339TimePtr{Time: &closedAt},
	}
}

// createDoingTicket creates a ticket with doing status
func createDoingTicket(id string, startedAt time.Time) ticket.Ticket {
	return ticket.Ticket{
		ID:        id,
		StartedAt: ticket.RFC3339TimePtr{Time: &startedAt},
	}
}

// testFixture holds common test dependencies
type testFixture struct {
	app         *App
	mockGit     *mocks.MockGitClient
	mockManager *mocks.MockTicketManager
	config      *config.Config
}

// newTestFixture creates a new test fixture with mocks
func newTestFixture(t *testing.T) *testFixture {
	mockGit := new(mocks.MockGitClient)
	mockManager := new(mocks.MockTicketManager)

	cfg := config.Default()
	cfg.Git.DefaultBranch = testDefaultBranch

	app := &App{
		Config:       cfg,
		Git:          mockGit,
		Manager:      mockManager,
		Output:       NewOutputWriter(nil, nil, FormatText),
		StatusWriter: NewStatusWriter(os.Stdout, FormatText),
	}

	return &testFixture{
		app:         app,
		mockGit:     mockGit,
		mockManager: mockManager,
		config:      cfg,
	}
}

// assertMocks verifies all mock expectations were met
func (f *testFixture) assertMocks(t *testing.T) {
	f.mockGit.AssertExpectations(t)
	f.mockManager.AssertExpectations(t)
}

// NewAppWithWorkingDir creates a new App instance with a specific working directory for testing
func NewAppWithWorkingDir(ctx context.Context, t *testing.T, workingDir string) (*App, error) {
	t.Helper()
	return NewAppWithOptions(ctx, WithWorkingDirectory(workingDir))
}

// ConfigureTestGit sets up git configuration for a test repository
// IMPORTANT: Always configures locally, never globally
func ConfigureTestGit(t *testing.T, repoPath string) {
	t.Helper()

	ctx := context.Background()
	gitOps := git.New(repoPath)

	// Configure user name locally
	if _, err := gitOps.Exec(ctx, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("Failed to configure git user.name: %v", err)
	}

	// Configure user email locally
	if _, err := gitOps.Exec(ctx, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("Failed to configure git user.email: %v", err)
	}
}

// NewTestOutputWriter creates an OutputWriter that captures output for testing
func NewTestOutputWriter() (*OutputWriter, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	writer := NewOutputWriter(stdout, stderr, FormatText)
	return writer, stdout, stderr
}
