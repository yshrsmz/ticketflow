package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yshrsmz/ticketflow/internal/ticket"
	"github.com/yshrsmz/ticketflow/internal/ui/components"
	"github.com/yshrsmz/ticketflow/internal/ui/styles"
)

// Action represents an action to take from the list view
type Action int

const (
	ActionNone Action = iota
	ActionViewDetail
	ActionNewTicket
	ActionStartTicket
	ActionRefresh
)

// TicketListModel represents the ticket list view
type TicketListModel struct {
	manager      *ticket.Manager
	tickets      []ticket.Ticket
	cursor       int
	selected     map[string]bool
	err          error
	action       Action
	statusFilter string
	width        int
	height       int
}

// NewTicketListModel creates a new ticket list model
func NewTicketListModel(manager *ticket.Manager) TicketListModel {
	return TicketListModel{
		manager:  manager,
		selected: make(map[string]bool),
		action:   ActionNone,
	}
}

// Init initializes the model
func (m TicketListModel) Init() tea.Cmd {
	return m.loadTickets()
}

// Update handles messages
func (m TicketListModel) Update(msg tea.Msg) (TicketListModel, tea.Cmd) {
	m.action = ActionNone

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.tickets)-1 {
				m.cursor++
			}

		case "g", "home":
			m.cursor = 0

		case "G", "end":
			if len(m.tickets) > 0 {
				m.cursor = len(m.tickets) - 1
			}

		case "enter":
			m.action = ActionViewDetail

		case "n":
			m.action = ActionNewTicket

		case "s":
			m.action = ActionStartTicket

		case "r":
			m.action = ActionRefresh
			return m, m.loadTickets()

		case " ":
			if len(m.tickets) > 0 && m.cursor < len(m.tickets) {
				id := m.tickets[m.cursor].ID
				m.selected[id] = !m.selected[id]
			}

		case "1", "2", "3":
			// Filter by priority
			if msg.String() == m.statusFilter {
				m.statusFilter = ""
			} else {
				m.statusFilter = msg.String()
			}
			return m, m.loadTickets()

		case "t", "d", "x":
			// Filter by status: t=todo, d=doing, x=done
			statusMap := map[string]string{
				"t": "todo",
				"d": "doing",
				"x": "done",
			}
			status := statusMap[msg.String()]
			if status == m.statusFilter {
				m.statusFilter = ""
			} else {
				m.statusFilter = status
			}
			return m, m.loadTickets()
		}

	case ticketsLoadedMsg:
		m.tickets = msg.tickets
		m.err = msg.err
		if m.cursor >= len(m.tickets) {
			m.cursor = len(m.tickets) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}

	case error:
		m.err = msg
	}

	return m, nil
}

// View renders the view
func (m TicketListModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n  %s\n", styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	if len(m.tickets) == 0 {
		emptyMsg := "No tickets found."
		if m.statusFilter != "" {
			emptyMsg = fmt.Sprintf("No %s tickets found.", m.statusFilter)
		}
		return fmt.Sprintf("\n  %s\n\n  Press 'n' to create a new ticket.\n", styles.InfoStyle.Render(emptyMsg))
	}

	var s strings.Builder

	// Title with filter info
	title := "Tickets"
	if m.statusFilter != "" {
		title = fmt.Sprintf("Tickets (%s)", m.statusFilter)
	}
	s.WriteString(styles.TitleStyle.Render(title))
	s.WriteString("\n\n")

	// Calculate column widths
	idWidth := 20
	statusWidth := 7
	priorityWidth := 3
	progressWidth := 8
	descWidth := m.width - idWidth - statusWidth - priorityWidth - progressWidth - 12 // padding and borders

	// Header
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %s",
		idWidth, "ID",
		statusWidth, "Status",
		priorityWidth, "Pri",
		progressWidth, "Progress",
		"Description")
	s.WriteString(styles.SubtitleStyle.Render(header))
	s.WriteString("\n")
	s.WriteString(strings.Repeat("─", m.width-4))
	s.WriteString("\n")

	// Ticket list
	visibleStart := 0
	visibleEnd := len(m.tickets)
	maxVisible := m.height - 10 // Leave room for header and help

	if len(m.tickets) > maxVisible {
		// Scroll to keep cursor visible
		if m.cursor < visibleStart {
			visibleStart = m.cursor
			visibleEnd = visibleStart + maxVisible
		} else if m.cursor >= visibleStart+maxVisible {
			visibleEnd = m.cursor + 1
			visibleStart = visibleEnd - maxVisible
		}
	}

	for i := visibleStart; i < visibleEnd && i < len(m.tickets); i++ {
		t := m.tickets[i]
		
		// Format row
		statusStyle := styles.GetStatusStyle(string(t.Status()))
		priorityStyle := styles.GetPriorityStyle(t.Priority)
		
		id := truncate(t.ID, idWidth)
		status := statusStyle.Render(fmt.Sprintf("%-*s", statusWidth, t.Status()))
		priority := priorityStyle.Render(fmt.Sprintf("%d", t.Priority))
		
		// Format progress
		progressStr := fmt.Sprintf("%3d%%", t.Progress)
		if t.Status() == ticket.StatusDoing && t.Progress > 0 {
			if t.Progress >= 80 {
				progressStr = styles.SuccessStyle.Render(progressStr)
			} else if t.Progress >= 50 {
				progressStr = styles.WarningStyle.Render(progressStr)
			} else {
				progressStr = styles.InfoStyle.Render(progressStr)
			}
		} else {
			progressStr = styles.MutedStyle.Render(progressStr)
		}
		
		desc := truncate(t.Description, descWidth)

		row := fmt.Sprintf("%-*s %s %s %s %s",
			idWidth, id,
			status,
			priority,
			progressStr,
			desc)

		// Apply selection/cursor styling
		if i == m.cursor {
			row = styles.SelectedItemStyle.Render(row)
		} else if m.selected[t.ID] {
			row = styles.ActiveButtonStyle.Render(row)
		} else {
			row = styles.ItemStyle.Render(row)
		}

		s.WriteString(row)
		s.WriteString("\n")
	}

	// Scroll indicator
	if len(m.tickets) > maxVisible {
		s.WriteString("\n")
		scrollInfo := fmt.Sprintf("%d-%d of %d", visibleStart+1, visibleEnd, len(m.tickets))
		s.WriteString(styles.HelpStyle.Render(scrollInfo))
	}

	// Help line
	s.WriteString("\n\n")
	s.WriteString(components.ShortHelp())

	return s.String()
}

// SetSize sets the view size
func (m *TicketListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Action returns the current action
func (m TicketListModel) Action() Action {
	return m.action
}

// SelectedTicket returns the currently selected ticket
func (m TicketListModel) SelectedTicket() *ticket.Ticket {
	if m.cursor >= 0 && m.cursor < len(m.tickets) {
		return &m.tickets[m.cursor]
	}
	return nil
}

// Refresh reloads the ticket list
func (m TicketListModel) Refresh() tea.Cmd {
	return m.loadTickets()
}

// ticketsLoadedMsg is sent when tickets are loaded
type ticketsLoadedMsg struct {
	tickets []ticket.Ticket
	err     error
}

// loadTickets loads tickets from the manager
func (m TicketListModel) loadTickets() tea.Cmd {
	return func() tea.Msg {
		tickets, err := m.manager.List(m.statusFilter)
		return ticketsLoadedMsg{
			tickets: tickets,
			err:     err,
		}
	}
}

// truncate truncates a string to a maximum width
func truncate(s string, maxWidth int) string {
	if len(s) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return "..."
	}
	return s[:maxWidth-3] + "..."
}