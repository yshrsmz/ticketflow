package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yshrsmz/ticketflow/internal/ticket"
	"github.com/yshrsmz/ticketflow/internal/ui/styles"
)

// NewTicketState represents the state of the new ticket view
type NewTicketState int

const (
	NewTicketStateInput NewTicketState = iota
	NewTicketStateCreating
	NewTicketStateCreated
	NewTicketStateCancelled
	NewTicketStateError
)

// NewTicketModel represents the new ticket creation view
type NewTicketModel struct {
	manager    *ticket.Manager
	state      NewTicketState
	err        error
	width      int
	height     int
	focusIndex int

	// Form inputs
	slugInput     textinput.Model
	priorityInput textinput.Model
	descArea      textarea.Model
	contentArea   textarea.Model
}

// NewNewTicketModel creates a new ticket creation model
func NewNewTicketModel(manager *ticket.Manager) NewTicketModel {
	// Slug input
	slugInput := textinput.New()
	slugInput.Placeholder = "feature-name"
	slugInput.Focus()
	slugInput.CharLimit = 50
	slugInput.Width = 50
	slugInput.Prompt = ""

	// Priority input
	priorityInput := textinput.New()
	priorityInput.Placeholder = "3"
	priorityInput.CharLimit = 1
	priorityInput.Width = 10
	priorityInput.Prompt = ""

	// Description textarea
	descArea := textarea.New()
	descArea.Placeholder = "Brief description of the ticket..."
	descArea.CharLimit = 200
	descArea.SetWidth(60)
	descArea.SetHeight(2)
	descArea.ShowLineNumbers = false

	// Content textarea
	contentArea := textarea.New()
	contentArea.Placeholder = "Detailed ticket content (markdown)..."
	contentArea.SetWidth(60)
	contentArea.SetHeight(10)
	contentArea.ShowLineNumbers = false

	return NewTicketModel{
		manager:       manager,
		state:         NewTicketStateInput,
		slugInput:     slugInput,
		priorityInput: priorityInput,
		descArea:      descArea,
		contentArea:   contentArea,
		focusIndex:    0,
	}
}

// Init initializes the model
func (m NewTicketModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m NewTicketModel) Update(msg tea.Msg) (NewTicketModel, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, 4) // Initialize with size 4 for the 4 input fields

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.state == NewTicketStateInput {
				m.state = NewTicketStateCancelled
			}
			return m, nil

		case "tab", "shift+tab":
			if m.state == NewTicketStateInput {
				// Cycle through inputs
				if msg.String() == "tab" {
					m.focusIndex = (m.focusIndex + 1) % 4
				} else {
					m.focusIndex = (m.focusIndex - 1 + 4) % 4
				}

				// Update focus
				m.updateFocus()
			}

		case "ctrl+s":
			if m.state == NewTicketStateInput {
				m.state = NewTicketStateCreating
				return m, m.createTicket()
			}
		}

	case ticketCreatedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = NewTicketStateError
		} else {
			m.state = NewTicketStateCreated
		}
		return m, nil

	case error:
		m.err = msg
		m.state = NewTicketStateError
		return m, nil
	}

	// Update the focused input
	if m.state == NewTicketStateInput {
		switch m.focusIndex {
		case 0:
			m.slugInput, cmds[0] = m.slugInput.Update(msg)
		case 1:
			m.priorityInput, cmds[1] = m.priorityInput.Update(msg)
		case 2:
			m.descArea, cmds[2] = m.descArea.Update(msg)
		case 3:
			m.contentArea, cmds[3] = m.contentArea.Update(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the view
func (m NewTicketModel) View() string {
	if m.state == NewTicketStateCreated {
		return fmt.Sprintf("\n  %s\n\n  Press any key to continue.",
			styles.SuccessStyle.Render("✓ Ticket created successfully!"))
	}

	if m.state == NewTicketStateCancelled {
		return ""
	}

	if m.state == NewTicketStateError {
		return fmt.Sprintf("\n  %s\n\n  Press esc to go back.",
			styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	if m.state == NewTicketStateCreating {
		return "\n  Creating ticket..."
	}

	var s strings.Builder

	// Title
	s.WriteString(styles.TitleStyle.Render("New Ticket"))
	s.WriteString("\n\n")

	// Form
	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.GetPriorityStyle(2).GetForeground()).
		Padding(1, 2).
		Width(m.width - 4)

	var form strings.Builder

	// Slug field
	form.WriteString(styles.SubtitleStyle.Render("Slug:"))
	form.WriteString("\n")
	if m.focusIndex == 0 {
		form.WriteString(styles.FocusedInputStyle.Render(m.slugInput.View()))
	} else {
		form.WriteString(styles.InputStyle.Render(m.slugInput.View()))
	}
	form.WriteString("\n\n")

	// Priority field
	form.WriteString(styles.SubtitleStyle.Render("Priority (1-3):"))
	form.WriteString("\n")
	if m.focusIndex == 1 {
		form.WriteString(styles.FocusedInputStyle.Render(m.priorityInput.View()))
	} else {
		form.WriteString(styles.InputStyle.Render(m.priorityInput.View()))
	}
	form.WriteString("\n\n")

	// Description field
	form.WriteString(styles.SubtitleStyle.Render("Description:"))
	form.WriteString("\n")
	if m.focusIndex == 2 {
		descStyle := styles.FocusedInputStyle.Copy().UnsetBorderStyle()
		form.WriteString(descStyle.Render(m.descArea.View()))
	} else {
		descStyle := styles.InputStyle.Copy().UnsetBorderStyle()
		form.WriteString(descStyle.Render(m.descArea.View()))
	}
	form.WriteString("\n\n")

	// Content field
	form.WriteString(styles.SubtitleStyle.Render("Content:"))
	form.WriteString("\n")
	if m.focusIndex == 3 {
		contentStyle := styles.FocusedInputStyle.Copy().UnsetBorderStyle()
		form.WriteString(contentStyle.Render(m.contentArea.View()))
	} else {
		contentStyle := styles.InputStyle.Copy().UnsetBorderStyle()
		form.WriteString(contentStyle.Render(m.contentArea.View()))
	}

	s.WriteString(formStyle.Render(form.String()))

	// Help
	s.WriteString("\n\n")
	help := []string{
		fmt.Sprintf("%s navigate", styles.HelpKeyStyle.Render("tab")),
		fmt.Sprintf("%s save", styles.HelpKeyStyle.Render("ctrl+s")),
		fmt.Sprintf("%s cancel", styles.HelpKeyStyle.Render("esc")),
	}
	s.WriteString(strings.Join(help, " • "))

	return s.String()
}

// SetSize sets the view size
func (m *NewTicketModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Update component sizes
	m.slugInput.Width = min(50, width-10)
	m.descArea.SetWidth(min(60, width-10))
	m.contentArea.SetWidth(min(60, width-10))

	// Adjust content area height based on available space
	availableHeight := height - 25 // Account for other UI elements
	m.contentArea.SetHeight(min(10, max(5, availableHeight)))
}

// Reset resets the form
func (m *NewTicketModel) Reset() {
	m.state = NewTicketStateInput
	m.err = nil
	m.focusIndex = 0

	m.slugInput.Reset()
	m.priorityInput.Reset()
	m.priorityInput.SetValue("3")
	m.descArea.Reset()
	m.contentArea.Reset()

	m.updateFocus()
}

// State returns the current state
func (m NewTicketModel) State() NewTicketState {
	return m.state
}

// updateFocus updates which input has focus
func (m *NewTicketModel) updateFocus() {
	m.slugInput.Blur()
	m.priorityInput.Blur()
	m.descArea.Blur()
	m.contentArea.Blur()

	switch m.focusIndex {
	case 0:
		m.slugInput.Focus()
	case 1:
		m.priorityInput.Focus()
	case 2:
		m.descArea.Focus()
	case 3:
		m.contentArea.Focus()
	}
}

// ticketCreatedMsg is sent when a ticket is created
type ticketCreatedMsg struct {
	err error
}

// createTicket creates a new ticket
func (m NewTicketModel) createTicket() tea.Cmd {
	return func() tea.Msg {
		slug := strings.TrimSpace(m.slugInput.Value())
		if slug == "" {
			return ticketCreatedMsg{err: fmt.Errorf("slug is required")}
		}

		// Parse priority
		priority := 3
		if p := strings.TrimSpace(m.priorityInput.Value()); p != "" {
			switch p {
			case "1":
				priority = 1
			case "2":
				priority = 2
			case "3":
				priority = 3
			default:
				return ticketCreatedMsg{err: fmt.Errorf("priority must be 1, 2, or 3")}
			}
		}

		// Create ticket
		t, err := m.manager.Create(slug)
		if err != nil {
			return ticketCreatedMsg{err: err}
		}

		// Update metadata
		t.Priority = priority
		t.Description = strings.TrimSpace(m.descArea.Value())

		// Save ticket
		err = m.manager.Update(t)
		if err != nil {
			return ticketCreatedMsg{err: err}
		}

		// Save content if provided
		if content := strings.TrimSpace(m.contentArea.Value()); content != "" {
			err = m.manager.WriteContent(t.ID, content)
			if err != nil {
				return ticketCreatedMsg{err: err}
			}
		}

		return ticketCreatedMsg{err: nil}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
