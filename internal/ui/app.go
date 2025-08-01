package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

// ticketStartedMsg is sent when a ticket is successfully started
type ticketStartedMsg struct {
	ticket       *ticket.Ticket
	worktreePath string
	initWarning  string // Warning message if init commands failed
}

// ticketClosedMsg is sent when a ticket is successfully closed
type ticketClosedMsg struct {
	ticket       *ticket.Ticket
	isWorktree   bool
	worktreePath string
}

// ticketEditedMsg is sent when a ticket has been edited
type ticketEditedMsg struct {
	ticket *ticket.Ticket
}

// Model represents the application state
type Model struct {
	// Core components
	config      *config.Config
	manager     ticket.TicketManager
	git         git.GitClient
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
func New(cfg *config.Config, manager ticket.TicketManager, gitClient git.GitClient, projectRoot string) Model {
	return Model{
		config:       cfg,
		manager:      manager,
		git:          gitClient,
		projectRoot:  projectRoot,
		view:         ViewTicketList,
		previousView: ViewTicketList,
		ticketList:   views.NewTicketListModel(manager),
		ticketDetail: views.NewTicketDetailModel(manager),
		newTicket:    views.NewNewTicketModel(manager),
		worktreeList: views.NewWorktreeListModel(gitClient, cfg),
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

		// Skip most global shortcuts when in text input views
		isInTextInputMode := m.view == ViewNewTicket || (m.view == ViewTicketList && m.ticketList.IsSearchMode())
		if isInTextInputMode {
			// Only handle ctrl+c for emergency exit
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			// Let all other keys pass to the view for text input
		} else {
			// Global shortcuts for non-text-input views
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

	case ticketStartedMsg:
		// Ticket was successfully started
		var successMsg string
		if msg.worktreePath != "" {
			successMsg = fmt.Sprintf("✅ Ticket started! Worktree created at: %s", msg.worktreePath)
		} else {
			successMsg = fmt.Sprintf("✅ Ticket started! Switched to branch: %s", msg.ticket.ID)
		}

		// Add warning if init commands failed
		if msg.initWarning != "" {
			successMsg += fmt.Sprintf("\n⚠️  Warning: %s", msg.initWarning)
		}

		m.err = fmt.Errorf("%s", successMsg)
		// Refresh the list
		cmds = append(cmds, m.ticketList.Refresh())
		return m, tea.Batch(cmds...)

	case ticketClosedMsg:
		// Ticket was successfully closed
		if msg.isWorktree {
			m.err = fmt.Errorf("✅ Ticket closed! Worktree remains at: %s", msg.worktreePath)
		} else {
			m.err = fmt.Errorf("✅ Ticket closed! Branch: %s", msg.ticket.ID)
		}
		// Go back to list and refresh
		if m.view == ViewTicketDetail {
			m.view = m.previousView
		}
		cmds = append(cmds, m.ticketList.Refresh())
		return m, tea.Batch(cmds...)

	case ticketEditedMsg:
		// Ticket was edited, update detail view if showing
		if m.view == ViewTicketDetail {
			m.ticketDetail.SetTicket(msg.ticket)
			cmds = append(cmds, m.ticketDetail.Init())
		}
		return m, tea.Batch(cmds...)
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

		// Handle detail view actions
		switch m.ticketDetail.Action() {
		case views.DetailActionClose:
			t := m.ticketDetail.SelectedTicket()
			if t != nil {
				cmds = append(cmds, m.closeTicket(t))
			}

		case views.DetailActionEdit:
			t := m.ticketDetail.SelectedTicket()
			if t != nil {
				cmds = append(cmds, m.editTicket(t))
			}

		case views.DetailActionStart:
			t := m.ticketDetail.SelectedTicket()
			if t != nil {
				cmds = append(cmds, m.startTicket(t))
			}
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
		// Validate ticket can be started
		if err := m.validateTicketForStart(t); err != nil {
			return err
		}

		// Check workspace state
		if err := m.checkWorkspaceForStart(); err != nil {
			return err
		}

		// Get current branch
		currentBranch, err := m.git.CurrentBranch()
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}

		// Setup branch or worktree
		worktreePath, initErr := m.setupTicketBranchOrWorktree(t)
		if initErr != nil && !strings.Contains(initErr.Error(), "initialization commands failed") {
			// If it's not an init command error, it's a fatal error
			return initErr
		}

		// Move ticket to doing status and commit
		if err := m.moveTicketToDoingAndCommit(t, worktreePath, currentBranch); err != nil {
			return err
		}

		// Return success message with any init warning
		msg := ticketStartedMsg{
			ticket:       t,
			worktreePath: worktreePath,
		}
		if initErr != nil {
			msg.initWarning = initErr.Error()
		}
		return msg
	}
}

// closeTicket closes a ticket
func (m *Model) closeTicket(t *ticket.Ticket) tea.Cmd {
	return func() tea.Msg {
		// Validate ticket can be closed
		if err := m.validateTicketForClose(t); err != nil {
			return err
		}

		// Check workspace state and get worktree info
		worktreePath, isWorktree, err := m.checkWorkspaceForClose(t)
		if err != nil {
			return err
		}

		// Move ticket to done status and commit
		if err := m.moveTicketToDoneAndCommit(t); err != nil {
			return err
		}

		return ticketClosedMsg{
			ticket:       t,
			isWorktree:   isWorktree,
			worktreePath: worktreePath,
		}
	}
}

// editTicket opens a ticket in the external editor
func (m *Model) editTicket(t *ticket.Ticket) tea.Cmd {
	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // fallback
	}

	// Create command
	cmd := exec.Command(editor, t.Path)

	// Use tea.ExecProcess to properly handle terminal state
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}

		// Reload ticket to get updated content
		updated, reloadErr := m.manager.Get(t.ID)
		if reloadErr != nil {
			return fmt.Errorf("failed to reload ticket: %w", reloadErr)
		}

		return ticketEditedMsg{
			ticket: updated,
		}
	})
}

// validateTicketForStart validates that a ticket can be started
func (m *Model) validateTicketForStart(t *ticket.Ticket) error {
	if t.Status() == ticket.StatusDoing {
		return fmt.Errorf("ticket %s is already in progress", t.ID)
	}
	return nil
}

// checkWorkspaceForStart checks if the workspace is ready to start a ticket
func (m *Model) checkWorkspaceForStart() error {
	// Check for uncommitted changes (only if not using worktrees)
	if !m.config.Worktree.Enabled {
		dirty, err := m.git.HasUncommittedChanges()
		if err != nil {
			return fmt.Errorf("failed to check git status: %w", err)
		}
		if dirty {
			return fmt.Errorf("uncommitted changes detected - please commit or stash before starting a ticket")
		}
	}
	return nil
}

// setupTicketBranchOrWorktree creates a branch or worktree for the ticket
func (m *Model) setupTicketBranchOrWorktree(t *ticket.Ticket) (string, error) {
	var worktreePath string

	if m.config.Worktree.Enabled {
		// Check if worktree already exists
		if exists, err := m.git.HasWorktree(t.ID); err != nil {
			return "", fmt.Errorf("failed to check worktree: %w", err)
		} else if exists {
			return "", fmt.Errorf("worktree for ticket %s already exists", t.ID)
		}

		// Create worktree
		baseDir := m.config.GetWorktreePath(m.projectRoot)
		worktreePath = filepath.Join(baseDir, t.ID)

		if err := m.git.AddWorktree(worktreePath, t.ID); err != nil {
			return "", fmt.Errorf("failed to create worktree: %w", err)
		}

		// Run init commands if configured
		if err := m.runWorktreeInitCommands(worktreePath); err != nil {
			// Non-fatal: store as warning to display later
			return worktreePath, err
		}
	} else {
		// Original behavior: create and checkout branch
		if err := m.git.CreateBranch(t.ID); err != nil {
			return "", fmt.Errorf("failed to create branch: %w", err)
		}
	}

	return worktreePath, nil
}

// runWorktreeInitCommands runs initialization commands in the worktree
func (m *Model) runWorktreeInitCommands(worktreePath string) error {
	if len(m.config.Worktree.InitCommands) == 0 {
		return nil
	}

	var failedCommands []string
	for _, cmd := range m.config.Worktree.InitCommands {
		parts := strings.Fields(cmd)
		if len(parts) == 0 {
			continue
		}

		execCmd := exec.Command(parts[0], parts[1:]...)
		execCmd.Dir = worktreePath
		if err := execCmd.Run(); err != nil {
			failedCommands = append(failedCommands, fmt.Sprintf("%s (%v)", cmd, err))
		}
	}

	if len(failedCommands) > 0 {
		return fmt.Errorf("some initialization commands failed: %s", strings.Join(failedCommands, ", "))
	}
	return nil
}

// moveTicketToDoingAndCommit moves ticket to doing status and commits the change
func (m *Model) moveTicketToDoingAndCommit(t *ticket.Ticket, worktreePath, currentBranch string) error {
	// Update ticket status
	if err := t.Start(); err != nil {
		// Rollback
		m.rollbackTicketStart(worktreePath, currentBranch)
		return fmt.Errorf("failed to start ticket: %w", err)
	}

	// Move ticket file from todo to doing
	oldPath := t.Path
	doingPath := m.config.GetDoingPath(m.projectRoot)
	newPath := filepath.Join(doingPath, filepath.Base(t.Path))

	// Move the file
	if err := os.Rename(oldPath, newPath); err != nil {
		// Rollback
		m.rollbackTicketStart(worktreePath, currentBranch)
		return fmt.Errorf("failed to move ticket to doing: %w", err)
	}

	// Update ticket data with new path
	t.Path = newPath
	if err := m.manager.Update(t); err != nil {
		// Rollback file move
		_ = os.Rename(newPath, oldPath)
		m.rollbackTicketStart(worktreePath, currentBranch)
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	// Git add both old and new paths
	if err := m.git.Add(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to stage ticket move: %w", err)
	}

	// Commit the move
	if err := m.git.Commit(fmt.Sprintf("Start ticket: %s", t.ID)); err != nil {
		return fmt.Errorf("failed to commit ticket move: %w", err)
	}

	// Set current ticket
	if err := m.manager.SetCurrentTicket(t); err != nil {
		return fmt.Errorf("failed to set current ticket: %w", err)
	}

	return nil
}

// rollbackTicketStart rolls back changes made during ticket start
func (m *Model) rollbackTicketStart(worktreePath, currentBranch string) {
	if m.config.Worktree.Enabled && worktreePath != "" {
		_ = m.git.RemoveWorktree(worktreePath)
	} else {
		_ = m.git.Checkout(currentBranch)
	}
}

// validateTicketForClose validates that a ticket can be closed
func (m *Model) validateTicketForClose(t *ticket.Ticket) error {
	// Get current ticket
	current, err := m.manager.GetCurrentTicket()
	if err != nil {
		return fmt.Errorf("failed to get current ticket: %w", err)
	}
	if current == nil || current.ID != t.ID {
		return fmt.Errorf("can only close the current active ticket")
	}
	return nil
}

// checkWorkspaceForClose checks workspace state and returns worktree info
func (m *Model) checkWorkspaceForClose(t *ticket.Ticket) (string, bool, error) {
	var worktreePath string
	var isWorktree bool

	if m.config.Worktree.Enabled {
		// Check if a worktree exists for this ticket
		wt, err := m.git.FindWorktreeByBranch(t.ID)
		if err != nil {
			return "", false, fmt.Errorf("failed to find worktree: %w", err)
		}
		if wt != nil {
			isWorktree = true
			worktreePath = wt.Path

			// Check for uncommitted changes in worktree
			wtGit := git.New(worktreePath)
			dirty, err := wtGit.HasUncommittedChanges()
			if err != nil {
				return "", false, fmt.Errorf("failed to check worktree status: %w", err)
			}
			if dirty {
				return "", false, fmt.Errorf("uncommitted changes in worktree - please commit before closing")
			}
		}
	}

	if !isWorktree {
		// Check for uncommitted changes
		dirty, err := m.git.HasUncommittedChanges()
		if err != nil {
			return "", false, fmt.Errorf("failed to check git status: %w", err)
		}
		if dirty {
			return "", false, fmt.Errorf("uncommitted changes - please commit before closing")
		}

		// Get current branch
		currentBranch, err := m.git.CurrentBranch()
		if err != nil {
			return "", false, fmt.Errorf("failed to get current branch: %w", err)
		}

		// Ensure we're on the ticket branch
		if currentBranch != t.ID {
			return "", false, fmt.Errorf("not on ticket branch, expected %s but on %s", t.ID, currentBranch)
		}
	}

	return worktreePath, isWorktree, nil
}

// moveTicketToDoneAndCommit moves ticket to done status and commits the change
func (m *Model) moveTicketToDoneAndCommit(t *ticket.Ticket) error {
	// Update ticket status
	if err := t.Close(); err != nil {
		return fmt.Errorf("failed to close ticket: %w", err)
	}

	// Move ticket file from doing to done
	oldPath := t.Path
	donePath := m.config.GetDonePath(m.projectRoot)
	newPath := filepath.Join(donePath, filepath.Base(t.Path))

	// Move the file
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to move ticket to done: %w", err)
	}

	// Update ticket data with new path
	t.Path = newPath
	if err := m.manager.Update(t); err != nil {
		// Rollback file move
		_ = os.Rename(newPath, oldPath)
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	// Git add both old and new paths
	if err := m.git.Add(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to stage ticket move: %w", err)
	}

	// Commit the move
	if err := m.git.Commit(fmt.Sprintf("Close ticket: %s", t.ID)); err != nil {
		return fmt.Errorf("failed to commit ticket move: %w", err)
	}

	// Remove current ticket link
	if err := m.manager.SetCurrentTicket(nil); err != nil {
		return fmt.Errorf("failed to remove current ticket link: %w", err)
	}

	return nil
}
