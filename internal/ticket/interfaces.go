package ticket

// TicketManager defines the interface for ticket operations
type TicketManager interface {
	// Create creates a new ticket in the todo directory
	Create(slug string) (*Ticket, error)
	
	// Get retrieves a ticket by ID
	Get(id string) (*Ticket, error)
	
	// List returns tickets based on the status filter
	List(statusFilter StatusFilter) ([]Ticket, error)
	
	// Update updates an existing ticket
	Update(ticket *Ticket) error
	
	// GetCurrentTicket returns the currently active ticket (in 'doing' status)
	GetCurrentTicket() (*Ticket, error)
	
	// SetCurrentTicket sets the ticket as the current active ticket
	SetCurrentTicket(ticket *Ticket) error
	
	// ReadContent reads the content of a ticket file
	ReadContent(id string) (string, error)
	
	// WriteContent writes content to a ticket file
	WriteContent(id string, content string) error
	
	// FindTicket finds a ticket by ID across all directories
	FindTicket(ticketID string) (string, error)
}