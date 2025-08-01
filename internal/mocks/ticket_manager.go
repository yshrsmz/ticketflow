package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// MockTicketManager is a mock implementation of ticket.TicketManager
type MockTicketManager struct {
	mock.Mock
}

// Create creates a new ticket in the todo directory
func (m *MockTicketManager) Create(slug string) (*ticket.Ticket, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ticket.Ticket), args.Error(1)
}

// Get retrieves a ticket by ID
func (m *MockTicketManager) Get(id string) (*ticket.Ticket, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ticket.Ticket), args.Error(1)
}

// List returns tickets based on the status filter
func (m *MockTicketManager) List(statusFilter ticket.StatusFilter) ([]ticket.Ticket, error) {
	args := m.Called(statusFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ticket.Ticket), args.Error(1)
}

// Update updates an existing ticket
func (m *MockTicketManager) Update(t *ticket.Ticket) error {
	args := m.Called(t)
	return args.Error(0)
}

// GetCurrentTicket returns the currently active ticket (in 'doing' status)
func (m *MockTicketManager) GetCurrentTicket() (*ticket.Ticket, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ticket.Ticket), args.Error(1)
}

// SetCurrentTicket sets the ticket as the current active ticket
func (m *MockTicketManager) SetCurrentTicket(t *ticket.Ticket) error {
	args := m.Called(t)
	return args.Error(0)
}

// ReadContent reads the content of a ticket file
func (m *MockTicketManager) ReadContent(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

// WriteContent writes content to a ticket file
func (m *MockTicketManager) WriteContent(id string, content string) error {
	args := m.Called(id, content)
	return args.Error(0)
}

// FindTicket finds a ticket by ID across all directories
func (m *MockTicketManager) FindTicket(ticketID string) (string, error) {
	args := m.Called(ticketID)
	return args.String(0), args.Error(1)
}
