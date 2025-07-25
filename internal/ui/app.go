package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ticket"
	"github.com/yshrsmz/ticketflow/internal/ui/components"
	"github.com/yshrsmz/ticketflow/internal/ui/styles"
	"github.com/yshrsmz/ticketflow/internal/ui/views"
)

// ViewType represents the current view
type ViewType int

const (
	ViewTicketList ViewType = iota
	ViewTicketDetail
	ViewNewTicket
	ViewWorktreeList
)

// Model represents the application state
type Model struct {
	// Core components
	config      *config.Config
	manager     *ticket.Manager
	git         *git.Git
	projectRoot string

	// View state
	view         ViewType
	previousView ViewType

	// Views
	ticketList   views.TicketListModel
	ticketDetail views.TicketDetailModel
	newTicket    views.NewTicketModel
	worktreeList views.WorktreeListModel

	// UI state
	help   components.HelpModel
	width  int
	height int
	err    error
	ready  bool
}

// New creates a new TUI application
func New(cfg *config.Config, manager *ticket.Manager, git *git.Git, projectRoot string) Model {
	return Model{
		config:       cfg,
		manager:      manager,
		git:          git,
		projectRoot:  projectRoot,
		view:         ViewTicketList,
		previousView: ViewTicketList,
		ticketList:   views.NewTicketListModel(manager),
		ticketDetail: views.NewTicketDetailModel(manager),
		newTicket:    views.NewNewTicketModel(manager),
		worktreeList: views.NewWorktreeListModel(git, cfg),
		help:         components.NewHelpModel(),
		ready:        false,
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.ticketList.Init(),
		tea.SetWindowTitle("TicketFlow"),
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Handle global keys first
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Help overlay takes precedence
		if m.help.IsVisible() {
			switch msg.String() {
			case "?", "esc", "q":
				m.help.Hide()
				return m, nil
			}
			return m, nil
		}

		// Global shortcuts
		switch msg.String() {
		case "?":
			m.help.Toggle()
			return m, nil

		case "ctrl+c":
			return m, tea.Quit

		case "q":
			if m.view == ViewTicketList {
				return m, tea.Quit
			}
			// Otherwise, go back
			m.view = m.previousView
			return m, nil

		case "w":
			if m.view != ViewWorktreeList {
				m.previousView = m.view
				m.view = ViewWorktreeList
				return m, m.worktreeList.Init()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update all views with new size
		m.ticketList.SetSize(msg.Width, msg.Height)
		m.ticketDetail.SetSize(msg.Width, msg.Height)
		m.newTicket.SetSize(msg.Width, msg.Height)
		m.worktreeList.SetSize(msg.Width, msg.Height)

	case error:
		m.err = msg
		return m, nil
	}

	// Delegate to current view
	switch m.view {
	case ViewTicketList:
		m.ticketList, cmd = m.ticketList.Update(msg)
		cmds = append(cmds, cmd)

		// Handle view transitions
		switch m.ticketList.Action() {
		case views.ActionViewDetail:
			if selected := m.ticketList.SelectedTicket(); selected != nil {
				m.previousView = m.view
				m.view = ViewTicketDetail
				m.ticketDetail.SetTicket(selected)
				cmds = append(cmds, m.ticketDetail.Init())
			}

		case views.ActionNewTicket:
			m.previousView = m.view
			m.view = ViewNewTicket
			m.newTicket.Reset()
			cmds = append(cmds, m.newTicket.Init())

		case views.ActionStartTicket:
			if selected := m.ticketList.SelectedTicket(); selected != nil {
				cmds = append(cmds, m.startTicket(selected))
			}
		}

	case ViewTicketDetail:
		m.ticketDetail, cmd = m.ticketDetail.Update(msg)
		cmds = append(cmds, cmd)

		if m.ticketDetail.ShouldGoBack() {
			m.view = m.previousView
			// Refresh list
			cmds = append(cmds, m.ticketList.Refresh())
		}

	case ViewNewTicket:
		m.newTicket, cmd = m.newTicket.Update(msg)
		cmds = append(cmds, cmd)

		switch m.newTicket.State() {
		case views.NewTicketStateCancelled:
			m.view = m.previousView

		case views.NewTicketStateCreated:
			m.view = m.previousView
			// Refresh list
			cmds = append(cmds, m.ticketList.Refresh())
		}

	case ViewWorktreeList:
		m.worktreeList, cmd = m.worktreeList.Update(msg)
		cmds = append(cmds, cmd)

		if m.worktreeList.ShouldGoBack() {
			m.view = m.previousView
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	if m.err != nil {
		return fmt.Sprintf("\n  %s\n\n  Press q to quit.", styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	// Main content
	var content string
	switch m.view {
	case ViewTicketList:
		content = m.ticketList.View()
	case ViewTicketDetail:
		content = m.ticketDetail.View()
	case ViewNewTicket:
		content = m.newTicket.View()
	case ViewWorktreeList:
		content = m.worktreeList.View()
	}

	// Add help overlay if visible
	if m.help.IsVisible() {
		helpView := m.help.View()
		// Center the help overlay
		helpWidth := lipgloss.Width(helpView)
		helpHeight := lipgloss.Height(helpView)
		x := (m.width - helpWidth) / 2
		y := (m.height - helpHeight) / 2

		// Overlay help on content
		content = lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content)
		helpOverlay := lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top,
			lipgloss.NewStyle().Margin(y, 0, 0, x).Render(helpView))
		
		// Combine layers
		content = lipgloss.JoinVertical(lipgloss.Left, content[:y]) + 
			"\n" + helpOverlay[y:]
	}

	return content
}

// startTicket starts work on a ticket
func (m *Model) startTicket(t *ticket.Ticket) tea.Cmd {
	return func() tea.Msg {
		// Start ticket logic would go here
		// For now, just return an error message
		return fmt.Errorf("start ticket not implemented in TUI yet")
	}
}