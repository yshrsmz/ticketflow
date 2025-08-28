package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yshrsmz/ticketflow/internal/ui/styles"
)

// CloseDialogState represents the state of the close dialog
type CloseDialogState int

const (
	CloseDialogHidden CloseDialogState = iota
	CloseDialogInput
	CloseDialogConfirmed
	CloseDialogCancelled
)

// Dialog responsive behavior constants
const (
	// dialogWidthBreakpoint is the screen width below which the dialog adjusts its width
	dialogWidthBreakpoint = 75
	// dialogMargin is the margin to leave on each side when adjusting dialog width
	dialogMargin = 10
)

// CloseDialogModel represents the close dialog component
type CloseDialogModel struct {
	state         CloseDialogState
	reasonInput   textinput.Model
	width         int
	height        int
	requireReason bool // Whether reason is required (branch not merged)
	showError     bool
	errorMsg      string
}

// NewCloseDialogModel creates a new close dialog model
func NewCloseDialogModel() CloseDialogModel {
	ti := textinput.New()
	ti.Placeholder = "Enter reason for closing (e.g., requirements changed, duplicate, etc.)"
	ti.CharLimit = 200
	ti.Width = 60

	return CloseDialogModel{
		state:       CloseDialogHidden,
		reasonInput: ti,
	}
}

// Show displays the dialog
func (m *CloseDialogModel) Show(requireReason bool) {
	m.state = CloseDialogInput
	m.requireReason = requireReason
	m.reasonInput.Reset()
	m.reasonInput.Focus()
	m.showError = false
	m.errorMsg = ""
}

// Hide hides the dialog
func (m *CloseDialogModel) Hide() {
	m.state = CloseDialogHidden
	m.reasonInput.Blur()
	m.reasonInput.Reset()
	m.showError = false
}

// IsVisible returns whether the dialog is visible
func (m *CloseDialogModel) IsVisible() bool {
	return m.state == CloseDialogInput
}

// IsConfirmed returns whether the dialog was confirmed
func (m *CloseDialogModel) IsConfirmed() bool {
	return m.state == CloseDialogConfirmed
}

// IsCancelled returns whether the dialog was cancelled
func (m *CloseDialogModel) IsCancelled() bool {
	return m.state == CloseDialogCancelled
}

// GetReason returns the entered reason
func (m *CloseDialogModel) GetReason() string {
	return strings.TrimSpace(m.reasonInput.Value())
}

// SetSize updates the dialog dimensions
func (m *CloseDialogModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the model
func (m *CloseDialogModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m *CloseDialogModel) Update(msg tea.Msg) (*CloseDialogModel, tea.Cmd) {
	var cmd tea.Cmd

	if m.state != CloseDialogInput {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.state = CloseDialogCancelled
			m.Hide()
			return m, nil

		case "enter":
			reason := m.GetReason()

			// Validate input - reason is required when branch is not merged
			if m.requireReason {
				originalInput := m.reasonInput.Value()
				if originalInput != "" && reason == "" {
					m.showError = true
					m.errorMsg = "Reason cannot be only whitespace"
					return m, nil
				}
				if reason == "" {
					m.showError = true
					m.errorMsg = "Reason is required for unmerged branches"
					return m, nil
				}
			}

			m.state = CloseDialogConfirmed
			return m, nil

		default:
			// Clear error on typing
			if m.showError {
				m.showError = false
				m.errorMsg = ""
			}
		}
	}

	m.reasonInput, cmd = m.reasonInput.Update(msg)
	return m, cmd
}

// View renders the dialog
func (m *CloseDialogModel) View() string {
	if m.state != CloseDialogInput {
		return ""
	}

	// Build dialog content
	var content strings.Builder

	// Title
	title := "Close Ticket"
	if m.requireReason {
		title = "Close Ticket (Reason Required)"
	}
	content.WriteString(styles.TitleStyle.Render(title))
	content.WriteString("\n\n")

	// Instructions
	instructions := "Enter a reason for closing this ticket:"
	if !m.requireReason {
		instructions = "Enter an optional reason for closing this ticket (or press Enter to skip):"
	}
	content.WriteString(instructions)
	content.WriteString("\n\n")

	// Text input
	content.WriteString(m.reasonInput.View())

	// Error message
	if m.showError && m.errorMsg != "" {
		content.WriteString("\n\n")
		content.WriteString(styles.ErrorStyle.Render("⚠ " + m.errorMsg))
	}

	// Help text
	content.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().Faint(true)
	helpText := "Enter: Confirm • ESC: Cancel"
	content.WriteString(helpStyle.Render(helpText))

	// Apply dialog styling with dynamic width
	dialogContent := content.String()

	// Calculate appropriate width based on screen size
	dialogWidth := 65
	if m.width > 0 && m.width < dialogWidthBreakpoint {
		dialogWidth = m.width - dialogMargin // Leave some margin
	}

	return styles.DialogStyle.
		Width(dialogWidth).
		Padding(1, 2).
		Render(dialogContent)
}
