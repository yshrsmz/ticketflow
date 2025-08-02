package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yshrsmz/ticketflow/internal/ticket"
	"github.com/yshrsmz/ticketflow/internal/ui/styles"
)

const (
	// UI layout constants for content height calculation
	baseMetadataLines = 7  // Status, priority, created, description label, borders
	baseUIChrome      = 15 // Title, borders, padding, help text, spacing
)

// DetailAction represents an action from the detail view
type DetailAction int

const (
	DetailActionNone DetailAction = iota
	DetailActionClose
	DetailActionEdit
	DetailActionStart
)

// TicketDetailModel represents the ticket detail view
type TicketDetailModel struct {
	manager    ticket.TicketManager
	ticket     *ticket.Ticket
	content    string
	scrollY    int
	width      int
	height     int
	shouldBack bool
	action     DetailAction
	err        error
}

// NewTicketDetailModel creates a new ticket detail model
func NewTicketDetailModel(manager ticket.TicketManager) TicketDetailModel {
	return TicketDetailModel{
		manager: manager,
	}
}

// Init initializes the model
func (m TicketDetailModel) Init() tea.Cmd {
	if m.ticket != nil {
		return m.loadContent()
	}
	return nil
}

// Update handles messages
func (m TicketDetailModel) Update(msg tea.Msg) (TicketDetailModel, tea.Cmd) {
	m.shouldBack = false
	m.action = DetailActionNone

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.shouldBack = true

		case "c":
			if m.ticket != nil && m.ticket.Status() == ticket.StatusDoing {
				m.action = DetailActionClose
			}

		case "e":
			if m.ticket != nil {
				m.action = DetailActionEdit
			}

		case "s":
			if m.ticket != nil && m.ticket.Status() == ticket.StatusTodo {
				m.action = DetailActionStart
			}

		case "up", "k":
			if m.scrollY > 0 {
				m.scrollY--
			}

		case "down", "j":
			maxScroll := m.getMaxScroll()
			if m.scrollY < maxScroll {
				m.scrollY++
			}

		case "pgup":
			m.scrollY -= m.height / 2
			if m.scrollY < 0 {
				m.scrollY = 0
			}

		case "pgdown":
			m.scrollY += m.height / 2
			maxScroll := m.getMaxScroll()
			if m.scrollY > maxScroll {
				m.scrollY = maxScroll
			}

		case "g", "home":
			m.scrollY = 0

		case "G", "end":
			m.scrollY = m.getMaxScroll()
		}

	case contentLoadedMsg:
		m.content = msg.content
		m.err = msg.err

	case error:
		m.err = msg
	}

	return m, nil
}

// View renders the view
func (m TicketDetailModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n  %s\n", styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	if m.ticket == nil {
		return "\n  No ticket selected.\n"
	}

	var s strings.Builder

	// Title
	s.WriteString(styles.TitleStyle.Render(fmt.Sprintf("Ticket: %s", m.ticket.ID)))
	s.WriteString("\n\n")

	// Metadata section
	metaStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.GetStatusStyle(string(m.ticket.Status())).GetForeground()).
		Padding(1, 2).
		Width(m.width - 4)

	var meta strings.Builder
	meta.WriteString(fmt.Sprintf("%s %s\n",
		styles.SubtitleStyle.Render("Status:"),
		styles.GetStatusStyle(string(m.ticket.Status())).Render(string(m.ticket.Status()))))

	meta.WriteString(fmt.Sprintf("%s %s\n",
		styles.SubtitleStyle.Render("Priority:"),
		styles.GetPriorityStyle(m.ticket.Priority).Render(fmt.Sprintf("%d", m.ticket.Priority))))

	meta.WriteString(fmt.Sprintf("%s %s\n",
		styles.SubtitleStyle.Render("Created:"),
		styles.InfoStyle.Render(m.ticket.CreatedAt.Format(time.RFC3339))))

	if m.ticket.StartedAt.Time != nil {
		meta.WriteString(fmt.Sprintf("%s %s\n",
			styles.SubtitleStyle.Render("Started:"),
			styles.InfoStyle.Render(m.ticket.StartedAt.Time.Format(time.RFC3339))))
	}

	if m.ticket.ClosedAt.Time != nil {
		meta.WriteString(fmt.Sprintf("%s %s\n",
			styles.SubtitleStyle.Render("Closed:"),
			styles.InfoStyle.Render(m.ticket.ClosedAt.Time.Format(time.RFC3339))))
	}

	if len(m.ticket.Related) > 0 {
		meta.WriteString(fmt.Sprintf("%s %s\n",
			styles.SubtitleStyle.Render("Related:"),
			styles.InfoStyle.Render(strings.Join(m.ticket.Related, ", "))))
	}

	meta.WriteString(fmt.Sprintf("\n%s\n%s",
		styles.SubtitleStyle.Render("Description:"),
		lipgloss.NewStyle().Width(m.width-10).Render(m.ticket.Description)))

	s.WriteString(metaStyle.Render(meta.String()))
	s.WriteString("\n\n")

	// Content section
	if m.content != "" {
		contentHeight := m.getContentHeight()

		s.WriteString(styles.SubtitleStyle.Render("Content:"))
		s.WriteString("\n")

		// Apply scrolling
		lines := strings.Split(m.content, "\n")
		visibleLines := lines

		if len(lines) > contentHeight {
			start := m.scrollY
			end := start + contentHeight
			if end > len(lines) {
				end = len(lines)
			}
			visibleLines = lines[start:end]
		}

		contentBox := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.GetPriorityStyle(3).GetForeground()).
			Padding(0, 1).
			Width(m.width - 4).
			Height(contentHeight)

		s.WriteString(contentBox.Render(strings.Join(visibleLines, "\n")))

		// Scroll indicator
		if len(lines) > contentHeight {
			scrollInfo := fmt.Sprintf("Lines %d-%d of %d (↑/↓ to scroll)",
				m.scrollY+1,
				m.scrollY+len(visibleLines),
				len(lines))
			s.WriteString("\n")
			s.WriteString(styles.HelpStyle.Render(scrollInfo))
		}
	}

	// Help
	s.WriteString("\n\n")
	helpItems := []string{"q/esc: back"}
	if m.ticket != nil {
		if m.ticket.Status() == ticket.StatusTodo {
			helpItems = append(helpItems, "s: start")
		} else if m.ticket.Status() == ticket.StatusDoing {
			helpItems = append(helpItems, "c: close")
		}
		helpItems = append(helpItems, "e: edit")
	}
	if m.content != "" && strings.Count(m.content, "\n")+1 > m.getContentHeight() {
		helpItems = append(helpItems, "↑/↓/j/k: scroll", "g/G: top/bottom")
	}
	s.WriteString(styles.HelpStyle.Render(strings.Join(helpItems, " • ")))

	return s.String()
}

// SetSize sets the view size
func (m *TicketDetailModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetTicket sets the ticket to display
func (m *TicketDetailModel) SetTicket(t *ticket.Ticket) {
	m.ticket = t
	m.content = ""
	m.scrollY = 0
	m.err = nil
}

// ShouldGoBack returns whether the view should go back
func (m TicketDetailModel) ShouldGoBack() bool {
	return m.shouldBack
}

// Action returns the current action
func (m TicketDetailModel) Action() DetailAction {
	return m.action
}

// SelectedTicket returns the current ticket
func (m TicketDetailModel) SelectedTicket() *ticket.Ticket {
	return m.ticket
}

// getMaxScroll calculates the maximum scroll position
func (m TicketDetailModel) getMaxScroll() int {
	lines := strings.Count(m.content, "\n") + 1
	contentHeight := m.getContentHeight()
	maxScroll := lines - contentHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// getContentHeight calculates the available height for content display
func (m TicketDetailModel) getContentHeight() int {
	// Calculate metadata section height
	metaLines := baseMetadataLines
	if m.ticket != nil {
		if m.ticket.StartedAt.Time != nil {
			metaLines++
		}
		if m.ticket.ClosedAt.Time != nil {
			metaLines++
		}
		if len(m.ticket.Related) > 0 {
			metaLines++
		}
		// Add lines for description wrapping
		descWidth := m.width - 10
		if descWidth > 0 {
			descLines := (len(m.ticket.Description) + descWidth - 1) / descWidth
			metaLines += descLines
		}
	}

	// Account for UI chrome: title, borders, padding, help text
	uiChrome := baseUIChrome + metaLines
	contentHeight := m.height - uiChrome
	if contentHeight < 1 {
		contentHeight = 1
	}
	return contentHeight
}

// contentLoadedMsg is sent when content is loaded
type contentLoadedMsg struct {
	content string
	err     error
}

// loadContent loads the ticket content
func (m TicketDetailModel) loadContent() tea.Cmd {
	return func() tea.Msg {
		if m.ticket == nil {
			return contentLoadedMsg{err: fmt.Errorf("no ticket selected")}
		}

		content, err := m.manager.ReadContent(context.Background(), m.ticket.ID)
		return contentLoadedMsg{
			content: content,
			err:     err,
		}
	}
}
