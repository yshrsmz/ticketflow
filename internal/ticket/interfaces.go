package ticket

import "context"

// TicketManager defines the interface for ticket operations
type TicketManager interface {
	// Create creates a new ticket in the todo directory
	Create(ctx context.Context, slug string) (*Ticket, error)

	// Get retrieves a ticket by ID
	Get(ctx context.Context, id string) (*Ticket, error)

	// List returns tickets based on the status filter
	List(ctx context.Context, statusFilter StatusFilter) ([]Ticket, error)

	// Update updates an existing ticket
	Update(ctx context.Context, ticket *Ticket) error

	// GetCurrentTicket returns the currently active ticket (in 'doing' status)
	GetCurrentTicket(ctx context.Context) (*Ticket, error)

	// SetCurrentTicket sets the ticket as the current active ticket
	SetCurrentTicket(ctx context.Context, ticket *Ticket) error

	// ReadContent reads the content of a ticket file
	ReadContent(ctx context.Context, id string) (string, error)

	// WriteContent writes content to a ticket file
	WriteContent(ctx context.Context, id string, content string) error

	// FindTicket finds a ticket by ID across all directories
	FindTicket(ctx context.Context, ticketID string) (string, error)
}
