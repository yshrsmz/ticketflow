package testutil

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/mocks"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// MockSetup contains commonly used mocks and test infrastructure
type MockSetup struct {
	Git         *mocks.MockGitClient
	Manager     *mocks.MockTicketManager
	Output      *OutputCapture
	Config      *config.Config
	ProjectRoot string
}

// NewMockSetup creates a new mock setup with all commonly used test infrastructure
func NewMockSetup(t *testing.T, tmpDir string, opts ...MockOption) *MockSetup {
	t.Helper()

	// Create mocks
	mockGit := new(mocks.MockGitClient)
	mockManager := new(mocks.MockTicketManager)

	// Create output capture
	output := NewOutputCapture()

	// Create config
	cfg := ConfigFixture()

	setup := &MockSetup{
		Git:         mockGit,
		Manager:     mockManager,
		Output:      output,
		Config:      cfg,
		ProjectRoot: tmpDir,
	}

	// Apply options
	for _, opt := range opts {
		opt(setup)
	}

	return setup
}

// MockOption customizes the mock setup
type MockOption func(*MockSetup)

// WithJSONOutput configures the setup for JSON output
func WithJSONOutput() MockOption {
	return func(s *MockSetup) {
		s.Config.Output.DefaultFormat = "json"
	}
}

// WithConfig sets a custom config
func WithConfig(cfg *config.Config) MockOption {
	return func(s *MockSetup) {
		s.Config = cfg
	}
}

// AssertExpectations asserts all mock expectations
func (s *MockSetup) AssertExpectations(t *testing.T) {
	t.Helper()
	s.Git.AssertExpectations(t)
	s.Manager.AssertExpectations(t)
}

// AssertNoOutput asserts that no output was written
func (s *MockSetup) AssertNoOutput(t *testing.T) {
	t.Helper()
	if s.Output.Stdout() != "" {
		t.Errorf("Expected no stdout, got: %s", s.Output.Stdout())
	}
	if s.Output.Stderr() != "" {
		t.Errorf("Expected no stderr, got: %s", s.Output.Stderr())
	}
}

// ExpectGitWorkDir sets up expectation for WorkDir call
func (s *MockSetup) ExpectGitWorkDir(dir string) {
	s.Git.On("WorkDir").Return(dir).Once()
}

// ExpectGitCurrentBranch sets up expectation for CurrentBranch call
func (s *MockSetup) ExpectGitCurrentBranch(branch string, err error) {
	s.Git.On("CurrentBranch").Return(branch, err).Once()
}

// ExpectGitBranchExists sets up expectation for BranchExists call
func (s *MockSetup) ExpectGitBranchExists(branch string, exists bool, err error) {
	s.Git.On("BranchExists", branch).Return(exists, err).Once()
}

// ExpectGitCheckout sets up expectation for Checkout call
func (s *MockSetup) ExpectGitCheckout(branch string, create bool, err error) {
	s.Git.On("Checkout", branch, create).Return(err).Once()
}

// ExpectManagerList sets up expectation for List call
func (s *MockSetup) ExpectManagerList(tickets []*ticket.Ticket, err error) {
	s.Manager.On("List", mock.Anything).Return(tickets, err).Once()
}

// ExpectManagerGet sets up expectation for Get call
func (s *MockSetup) ExpectManagerGet(id string, ticket *ticket.Ticket, err error) {
	s.Manager.On("Get", mock.Anything, id).Return(ticket, err).Once()
}

// ExpectManagerCreate sets up expectation for Create call
func (s *MockSetup) ExpectManagerCreate(ticket *ticket.Ticket, err error) {
	s.Manager.On("Create", mock.Anything, ticket).Return(err).Once()
}

// ExpectManagerUpdate sets up expectation for Update call
func (s *MockSetup) ExpectManagerUpdate(ticket *ticket.Ticket, err error) {
	s.Manager.On("Update", mock.Anything, ticket).Return(err).Once()
}

// ExpectManagerDelete sets up expectation for Delete call
func (s *MockSetup) ExpectManagerDelete(id string, err error) {
	s.Manager.On("Delete", mock.Anything, id).Return(err).Once()
}

// StandardGitExpectations sets up common git expectations
func StandardGitExpectations(mockGit *mocks.MockGitClient, workDir, currentBranch string) {
	mockGit.On("WorkDir").Return(workDir).Maybe()
	mockGit.On("CurrentBranch").Return(currentBranch, nil).Maybe()
}
