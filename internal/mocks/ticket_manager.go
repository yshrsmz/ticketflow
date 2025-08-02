package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// MockTicketManager is a mock implementation of ticket.TicketManager
type MockTicketManager struct {
	mock.Mock
}

// Create creates a new ticket in the todo directory
func (m *MockTicketManager) Create(ctx context.Context, slug string) (*ticket.Ticket, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ticket.Ticket), args.Error(1)
}

// Get retrieves a ticket by ID
func (m *MockTicketManager) Get(ctx context.Context, id string) (*ticket.Ticket, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ticket.Ticket), args.Error(1)
}

// List returns tickets based on the status filter
func (m *MockTicketManager) List(ctx context.Context, statusFilter ticket.StatusFilter) ([]ticket.Ticket, error) {
	args := m.Called(ctx, statusFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ticket.Ticket), args.Error(1)
}

// Update updates an existing ticket
func (m *MockTicketManager) Update(ctx context.Context, t *ticket.Ticket) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

// GetCurrentTicket returns the currently active ticket (in 'doing' status)
func (m *MockTicketManager) GetCurrentTicket(ctx context.Context) (*ticket.Ticket, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ticket.Ticket), args.Error(1)
}

// SetCurrentTicket sets the ticket as the current active ticket
func (m *MockTicketManager) SetCurrentTicket(ctx context.Context, t *ticket.Ticket) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

// ReadContent reads the content of a ticket file
func (m *MockTicketManager) ReadContent(ctx context.Context, id string) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

// WriteContent writes content to a ticket file
func (m *MockTicketManager) WriteContent(ctx context.Context, id string, content string) error {
	args := m.Called(ctx, id, content)
	return args.Error(0)
}

// FindTicket finds a ticket by ID across all directories
func (m *MockTicketManager) FindTicket(ctx context.Context, ticketID string) (string, error) {
	args := m.Called(ctx, ticketID)
	return args.String(0), args.Error(1)
}
