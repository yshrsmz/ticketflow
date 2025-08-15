package commands

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/config"
	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
	"github.com/yshrsmz/ticketflow/internal/mocks"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// MockAppFactory is a function type for creating mock App instances
type MockAppFactory func(ctx context.Context) (*cli.App, error)

// testAppFactory is a package-level variable that can be overridden in tests
var testAppFactory MockAppFactory

// SetTestAppFactory sets a mock App factory for testing
func SetTestAppFactory(factory MockAppFactory) {
	testAppFactory = factory
}

// ResetTestAppFactory resets the App factory to default
func ResetTestAppFactory() {
	testAppFactory = nil
}

// TestFixture holds common test dependencies for command tests
type TestFixture struct {
	App         *cli.App
	MockGit     *mocks.MockGitClient
	MockManager *mocks.MockTicketManager
	Config      *config.Config
	Stdout      *bytes.Buffer
	Stderr      *bytes.Buffer
}

// NewTestFixture creates a new test fixture with mocked dependencies
func NewTestFixture(t *testing.T) *TestFixture {
	t.Helper()

	mockGit := new(mocks.MockGitClient)
	mockManager := new(mocks.MockTicketManager)

	cfg := config.Default()
	cfg.Git.DefaultBranch = "main"
	cfg.Worktree.Enabled = true
	cfg.Worktree.BaseDir = "../test-worktrees"

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	app := &cli.App{
		Config:  cfg,
		Git:     mockGit,
		Manager: mockManager,
		Output:  cli.NewOutputWriter(stdout, stderr, cli.FormatText),
	}

	return &TestFixture{
		App:         app,
		MockGit:     mockGit,
		MockManager: mockManager,
		Config:      cfg,
		Stdout:      stdout,
		Stderr:      stderr,
	}
}

// SetOutputFormat changes the output format of the fixture's App
func (f *TestFixture) SetOutputFormat(format cli.OutputFormat) {
	f.App.Output = cli.NewOutputWriter(f.Stdout, f.Stderr, format)
}

// AssertMocks verifies all mock expectations were met
func (f *TestFixture) AssertMocks(t *testing.T) {
	f.MockGit.AssertExpectations(t)
	f.MockManager.AssertExpectations(t)
}

// CreateTestTicket creates a test ticket with common fields
func CreateTestTicket(id string, status ticket.Status) *ticket.Ticket {
	now := time.Now()
	t := &ticket.Ticket{
		ID:          id,
		Priority:    1,
		Description: "Test ticket",
		CreatedAt:   ticket.RFC3339Time{Time: now},
	}

	switch status {
	case ticket.StatusDoing:
		t.StartedAt = ticket.RFC3339TimePtr{Time: &now}
	case ticket.StatusDone:
		t.StartedAt = ticket.RFC3339TimePtr{Time: &now}
		t.ClosedAt = ticket.RFC3339TimePtr{Time: &now}
	}

	return t
}

// MockStartResult creates a mock StartTicketResult for testing
func MockStartResult(ticketID string) *cli.StartTicketResult {
	t := CreateTestTicket(ticketID, ticket.StatusDoing)
	return &cli.StartTicketResult{
		Ticket:               t,
		WorktreePath:         "../test-worktrees/" + ticketID,
		ParentBranch:         "main",
		InitCommandsExecuted: true,
	}
}

// SetupMockForClose sets up common mocks for close command tests
func SetupMockForClose(f *TestFixture, ticketID string, exists bool, status ticket.Status) {
	if exists {
		t := CreateTestTicket(ticketID, status)
		f.MockManager.On("GetTicket", ticketID).Return(t, nil).Maybe()
		f.MockManager.On("GetTicketByStatus", ticketID, mock.Anything).Return(t, nil).Maybe()
	} else {
		f.MockManager.On("GetTicket", ticketID).Return(nil, ticketerrors.ErrTicketNotFound).Maybe()
		f.MockManager.On("GetTicketByStatus", ticketID, mock.Anything).Return(nil, ticketerrors.ErrTicketNotFound).Maybe()
	}
}

// SetupMockForStart sets up common mocks for start command tests
func SetupMockForStart(f *TestFixture, ticketID string, exists bool, status ticket.Status) {
	if exists {
		t := CreateTestTicket(ticketID, status)
		f.MockManager.On("GetTicket", ticketID).Return(t, nil).Maybe()

		if status == ticket.StatusTodo {
			// Mock moving to doing
			f.MockManager.On("MoveTicket", ticketID, ticket.StatusTodo, ticket.StatusDoing).Return(nil).Maybe()
			// Mock setting start time
			f.MockManager.On("UpdateTicket", ticketID, mock.AnythingOfType("func(*ticket.Ticket) error")).Return(nil).Maybe()
		}
	} else {
		f.MockManager.On("GetTicket", ticketID).Return(nil, ticketerrors.ErrTicketNotFound).Maybe()
	}
}
