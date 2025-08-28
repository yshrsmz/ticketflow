package views

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yshrsmz/ticketflow/internal/ticket"
	"github.com/yshrsmz/ticketflow/internal/ui/components"
	"github.com/yshrsmz/ticketflow/internal/ui/styles"
)

// Constants for column width calculations
const (
	idColumnWidthPercentage = 0.30
	minIDColumnWidth        = 20
	maxIDColumnWidth        = 40
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
	manager         ticket.TicketManager
	tickets         []ticket.Ticket
	filteredTickets []ticket.Ticket
	cursor          int
	selected        map[string]bool
	err             error
	action          Action
	statusFilter    ticket.StatusFilter
	activeTab       int // 0=ALL, 1=TODO, 2=DOING, 3=DONE
	searchMode      bool
	searchQuery     string
	width           int
	height          int
}

// NewTicketListModel creates a new ticket list model
func NewTicketListModel(manager ticket.TicketManager) TicketListModel {
	return TicketListModel{
		manager:         manager,
		selected:        make(map[string]bool),
		action:          ActionNone,
		activeTab:       0, // Start with ALL tab
		filteredTickets: []ticket.Ticket{},
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
		// Handle search mode input
		if m.searchMode {
			switch msg.String() {
			case "enter":
				m.searchMode = false
				m.applyFilter()
			case "esc":
				m.searchMode = false
				m.searchQuery = ""
				m.applyFilter()
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.applyFilter()
				}
			default:
				if len(msg.String()) == 1 {
					m.searchQuery += msg.String()
					m.applyFilter()
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "/":
			m.searchMode = true
			m.searchQuery = ""
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.filteredTickets)-1 {
				m.cursor++
			}

		case "g", "home":
			m.cursor = 0

		case "G", "end":
			if len(m.filteredTickets) > 0 {
				m.cursor = len(m.filteredTickets) - 1
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
			if len(m.filteredTickets) > 0 && m.cursor < len(m.filteredTickets) {
				id := m.filteredTickets[m.cursor].ID
				m.selected[id] = !m.selected[id]
			}

		case "tab", "shift+tab":
			// Navigate tabs
			if msg.String() == "tab" {
				m.activeTab = (m.activeTab + 1) % 4
			} else {
				m.activeTab = (m.activeTab - 1 + 4) % 4
			}
			// Update status filter based on tab
			switch m.activeTab {
			case 0:
				m.statusFilter = ticket.StatusFilterActive
			case 1:
				m.statusFilter = ticket.StatusFilterTodo
			case 2:
				m.statusFilter = ticket.StatusFilterDoing
			case 3:
				m.statusFilter = ticket.StatusFilterDone
			}
			m.cursor = 0
			return m, m.loadTickets()

		case "1", "2", "3":
			// Jump to tab by number (1=TODO, 2=DOING, 3=DONE)
			tabMap := map[string]int{
				"1": 1,
				"2": 2,
				"3": 3,
			}
			if tab, ok := tabMap[msg.String()]; ok {
				m.activeTab = tab
				// Update status filter
				switch m.activeTab {
				case 1:
					m.statusFilter = ticket.StatusFilterTodo
				case 2:
					m.statusFilter = ticket.StatusFilterDoing
				case 3:
					m.statusFilter = ticket.StatusFilterDone
				}
				m.cursor = 0
				return m, m.loadTickets()
			}

		case "a":
			// Show all tickets
			m.activeTab = 0
			m.statusFilter = ticket.StatusFilterActive
			m.cursor = 0
			return m, m.loadTickets()
		}

	case ticketsLoadedMsg:
		m.tickets = msg.tickets
		m.err = msg.err
		m.applyFilter()
		if m.cursor >= len(m.filteredTickets) {
			m.cursor = len(m.filteredTickets) - 1
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

	var s strings.Builder

	// Tabs
	tabs := []string{"ALL", "TODO", "DOING", "DONE"}
	var tabBar strings.Builder
	for i, tab := range tabs {
		style := styles.ButtonStyle
		if i == m.activeTab {
			style = styles.ActiveButtonStyle
		}
		tabBar.WriteString(style.Render(tab))
		if i < len(tabs)-1 {
			tabBar.WriteString(" ")
		}
	}
	s.WriteString(tabBar.String())

	// Search bar
	if m.searchMode || m.searchQuery != "" {
		s.WriteString("\n\n")
		searchBar := fmt.Sprintf("🔍 Search: %s", m.searchQuery)
		if m.searchMode {
			searchBar += "_"
		}
		s.WriteString(styles.InputStyle.Render(searchBar))
		s.WriteString("\n")
	}

	s.WriteString("\n")

	// Calculate column widths
	// Dynamic ID width: 30% of terminal width with min 20, max 40
	idWidth := int(float64(m.width) * idColumnWidthPercentage)
	if idWidth < minIDColumnWidth {
		idWidth = minIDColumnWidth // Minimum width
	}
	if idWidth > maxIDColumnWidth {
		idWidth = maxIDColumnWidth // Maximum width
	}
	statusWidth := 7
	priorityWidth := 3
	descWidth := m.width - idWidth - statusWidth - priorityWidth - 8 // padding and borders

	// Header
	header := fmt.Sprintf("%-*s %-*s %-*s %s",
		idWidth, "ID",
		statusWidth, "Status",
		priorityWidth, "Pri",
		"Description")
	s.WriteString(styles.SubtitleStyle.Render(header))
	s.WriteString("\n")
	s.WriteString(strings.Repeat("─", m.width-4))
	s.WriteString("\n")

	// Handle empty state
	if len(m.filteredTickets) == 0 {
		emptyMsg := "No tickets found."
		if m.searchQuery != "" {
			emptyMsg = fmt.Sprintf("No tickets matching '%s'", m.searchQuery)
		} else if m.statusFilter != "" {
			emptyMsg = fmt.Sprintf("No %s tickets found.", m.statusFilter)
		}

		// Display empty message at the top-left, where tickets would normally appear
		styledEmptyMsg := styles.InfoStyle.Render(emptyMsg)
		s.WriteString(styledEmptyMsg)
		s.WriteString("\n\n")
		s.WriteString("Press 'n' to create a new ticket")
		s.WriteString("\n")

		// Fill remaining space to maintain consistent layout
		maxVisible := m.height - 10 // Leave room for header and help
		if m.searchMode || m.searchQuery != "" {
			maxVisible -= 3 // Account for search bar
		}
		contentHeight := 3 // Empty message takes about 3 lines
		remainingLines := maxVisible - contentHeight
		if remainingLines > 0 {
			s.WriteString(strings.Repeat("\n", remainingLines))
		}
	} else {
		// Ticket list
		visibleStart := 0
		visibleEnd := len(m.filteredTickets)
		maxVisible := m.height - 10 // Leave room for header and help
		if m.searchMode || m.searchQuery != "" {
			maxVisible -= 3 // Account for search bar
		}

		if len(m.filteredTickets) > maxVisible {
			// Scroll to keep cursor visible
			if m.cursor < visibleStart {
				visibleStart = m.cursor
				visibleEnd = visibleStart + maxVisible
			} else if m.cursor >= visibleStart+maxVisible {
				visibleEnd = m.cursor + 1
				visibleStart = visibleEnd - maxVisible
			}
		}

		for i := visibleStart; i < visibleEnd && i < len(m.filteredTickets); i++ {
			t := m.filteredTickets[i]

			// Format row
			statusStyle := styles.GetStatusStyle(string(t.Status()))
			priorityStyle := styles.GetPriorityStyle(t.Priority)

			// Add abandoned indicator for closed tickets with reason
			warningPrefix := ""
			if t.Status() == ticket.StatusDone && t.ClosureReason != "" {
				warningPrefix = "⚠ "
			}

			// Account for warning icon when truncating
			effectiveWidth := idWidth
			if warningPrefix != "" {
				effectiveWidth = idWidth - len(warningPrefix)
			}
			id := warningPrefix + truncate(t.ID, effectiveWidth)
			status := statusStyle.Render(fmt.Sprintf("%-*s", statusWidth, t.Status()))
			priority := priorityStyle.Render(fmt.Sprintf("%d", t.Priority))

			desc := truncate(t.Description, descWidth)

			row := fmt.Sprintf("%-*s %s %s %s",
				idWidth, id,
				status,
				priority,
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
		if len(m.filteredTickets) > maxVisible {
			s.WriteString("\n")
			scrollInfo := fmt.Sprintf("%d-%d of %d", visibleStart+1, visibleEnd, len(m.filteredTickets))
			s.WriteString(styles.HelpStyle.Render(scrollInfo))
		}
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
	if m.cursor >= 0 && m.cursor < len(m.filteredTickets) {
		return &m.filteredTickets[m.cursor]
	}
	return nil
}

// IsSearchMode returns true if the list is in search mode
func (m TicketListModel) IsSearchMode() bool {
	return m.searchMode
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
		tickets, err := m.manager.List(context.Background(), m.statusFilter)
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

// applyFilter applies the search query filter to tickets
func (m *TicketListModel) applyFilter() {
	// Pre-allocate filteredTickets with capacity of original tickets
	// In worst case, all tickets match the filter
	m.filteredTickets = make([]ticket.Ticket, 0, len(m.tickets))
	query := strings.ToLower(m.searchQuery)

	for _, t := range m.tickets {
		// If no search query, include all tickets
		if query == "" {
			m.filteredTickets = append(m.filteredTickets, t)
			continue
		}

		// Search in ID, description, and content
		if strings.Contains(strings.ToLower(t.ID), query) ||
			strings.Contains(strings.ToLower(t.Description), query) ||
			strings.Contains(strings.ToLower(t.Content), query) {
			m.filteredTickets = append(m.filteredTickets, t)
		}
	}
}
