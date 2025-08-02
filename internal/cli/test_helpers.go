package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/mocks"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// Test constants
const (
	testTicketID          = "250101-120000-test-feature"
	testDefaultBranch     = "main"
	orphanedWorktreeCount = 1
	staleBranchCount      = 2
)

// testTime parses a time string and fails the test if parsing fails
func testTime(t *testing.T, timeStr string) time.Time {
	parsed, err := time.Parse(time.RFC3339, timeStr)
	require.NoError(t, err)
	return parsed
}

// testTimePtr returns a pointer to a parsed time
func testTimePtr(t *testing.T, timeStr string) *time.Time {
	tm := testTime(t, timeStr)
	return &tm
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

// createTodoTicket creates a ticket with todo status
func createTodoTicket(id string) ticket.Ticket {
	return ticket.Ticket{
		ID: id,
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
		Config:  cfg,
		Git:     mockGit,
		Manager: mockManager,
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
