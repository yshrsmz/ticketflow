package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yshrsmz/ticketflow/internal/ui/styles"
)

// HelpItem represents a single help entry
type HelpItem struct {
	Key  string
	Desc string
}

// HelpModel represents the help overlay
type HelpModel struct {
	visible bool
	items   [][]HelpItem
}

// NewHelpModel creates a new help model
func NewHelpModel() HelpModel {
	return HelpModel{
		visible: false,
		items: [][]HelpItem{
			// Navigation
			{
				{Key: "↑/k", Desc: "Move up"},
				{Key: "↓/j", Desc: "Move down"},
				{Key: "g/home", Desc: "Go to top"},
				{Key: "G/end", Desc: "Go to bottom"},
			},
			// Actions
			{
				{Key: "enter", Desc: "Select/View"},
				{Key: "n", Desc: "New ticket"},
				{Key: "s", Desc: "Start ticket"},
				{Key: "c", Desc: "Close ticket"},
				{Key: "w", Desc: "Worktree view"},
			},
			// View controls
			{
				{Key: "tab", Desc: "Next tab"},
				{Key: "shift+tab", Desc: "Previous tab"},
				{Key: "1/2/3", Desc: "Jump to TODO/DOING/DONE"},
				{Key: "a", Desc: "Show all tickets"},
				{Key: "esc", Desc: "Back/Cancel"},
				{Key: "r", Desc: "Refresh"},
			},
			// General
			{
				{Key: "/", Desc: "Search"},
				{Key: "?", Desc: "Toggle help"},
				{Key: "q", Desc: "Quit"},
			},
		},
	}
}

// Toggle toggles the help visibility
func (m *HelpModel) Toggle() {
	m.visible = !m.visible
}

// IsVisible returns whether help is visible
func (m HelpModel) IsVisible() bool {
	return m.visible
}

// Hide hides the help
func (m *HelpModel) Hide() {
	m.visible = false
}

// View renders the help overlay
func (m HelpModel) View() string {
	if !m.visible {
		return ""
	}

	// Build help content
	var sections []string
	sectionTitles := []string{"Navigation", "Actions", "View Controls", "General"}

	for i, section := range m.items {
		var items []string
		maxKeyLen := 0

		// Find max key length for alignment
		for _, item := range section {
			if len(item.Key) > maxKeyLen {
				maxKeyLen = len(item.Key)
			}
		}

		// Build section items
		for _, item := range section {
			key := styles.HelpKeyStyle.Render(fmt.Sprintf("%-*s", maxKeyLen+2, item.Key))
			desc := styles.HelpStyle.Render(item.Desc)
			items = append(items, fmt.Sprintf("  %s %s", key, desc))
		}

		// Add section title
		title := styles.SubtitleStyle.Render(sectionTitles[i])
		sectionContent := lipgloss.JoinVertical(lipgloss.Left, title, strings.Join(items, "\n"))
		sections = append(sections, sectionContent)
	}

	// Join all sections
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	
	// Add title
	title := styles.TitleStyle.Render("Keyboard Shortcuts")
	fullContent := lipgloss.JoinVertical(lipgloss.Left, title, "", content)

	// Apply dialog styling
	return styles.DialogStyle.Render(fullContent)
}

// ShortHelp returns a one-line help string
func ShortHelp() string {
	items := []string{
		fmt.Sprintf("%s %s", styles.HelpKeyStyle.Render("?"), "help"),
		fmt.Sprintf("%s %s", styles.HelpKeyStyle.Render("n"), "new"),
		fmt.Sprintf("%s %s", styles.HelpKeyStyle.Render("s"), "start"),
		fmt.Sprintf("%s %s", styles.HelpKeyStyle.Render("q"), "quit"),
	}
	return strings.Join(items, " • ")
}