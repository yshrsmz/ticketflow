package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("#7D56F4")
	secondaryColor = lipgloss.Color("#F780E2")
	successColor   = lipgloss.Color("#04B575")
	warningColor   = lipgloss.Color("#ECFD65")
	errorColor     = lipgloss.Color("#FF6B6B")
	mutedColor     = lipgloss.Color("#626262")
	bgColor        = lipgloss.Color("#1a1a1a")
	fgColor        = lipgloss.Color("#FAFAFA")

	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Foreground(fgColor).
			Background(bgColor)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true)

	// List styles
	ListStyle = lipgloss.NewStyle().
			Margin(1, 0).
			Padding(0, 1)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(bgColor).
				Background(primaryColor).
				Bold(true)

	ItemStyle = lipgloss.NewStyle().
			Foreground(fgColor).
			PaddingLeft(2)

	// Status styles
	TodoStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Bold(true)

	DoingStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	DoneStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	// Priority styles
	Priority1Style = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	Priority2Style = lipgloss.NewStyle().
			Foreground(warningColor)

	Priority3Style = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Help styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// Border styles
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Dialog styles
	DialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			Margin(2, 4)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			Foreground(fgColor).
			Background(lipgloss.Color("#2a2a2a")).
			Padding(0, 1)

	FocusedInputStyle = InputStyle.Copy().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(primaryColor)

	// Button styles
	ButtonStyle = lipgloss.NewStyle().
			Foreground(bgColor).
			Background(mutedColor).
			Padding(0, 2).
			MarginRight(1)

	ActiveButtonStyle = ButtonStyle.Copy().
				Background(primaryColor).
				Bold(true)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	// Warning style
	WarningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	// Info style
	InfoStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// Spinner style
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	// Muted style
	MutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)
)

// GetStatusStyle returns the appropriate style for a status
func GetStatusStyle(status string) lipgloss.Style {
	switch status {
	case "todo":
		return TodoStyle
	case "doing":
		return DoingStyle
	case "done":
		return DoneStyle
	default:
		return BaseStyle
	}
}

// GetPriorityStyle returns the appropriate style for a priority
func GetPriorityStyle(priority int) lipgloss.Style {
	switch priority {
	case 1:
		return Priority1Style
	case 2:
		return Priority2Style
	default:
		return Priority3Style
	}
}
