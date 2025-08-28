package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-shellwords"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/log"
	"github.com/yshrsmz/ticketflow/internal/ticket"
	"github.com/yshrsmz/ticketflow/internal/ui/components"
	"github.com/yshrsmz/ticketflow/internal/ui/styles"
	"github.com/yshrsmz/ticketflow/internal/ui/views"
	"github.com/yshrsmz/ticketflow/internal/worktree"
)

// Operation timeout constants
const (
	closeOperationTimeout = 30 * time.Second
)

// InitCommandError represents a non-fatal error during worktree initialization
type InitCommandError struct {
	FailedCommands []string
	underlying     error
}

// Error implements the error interface
func (e *InitCommandError) Error() string {
	return fmt.Sprintf("some initialization commands failed: %s", strings.Join(e.FailedCommands, ", "))
}

// Unwrap implements the errors.Unwrap interface
func (e *InitCommandError) Unwrap() error {
	return e.underlying
}

// IsInitCommandError checks if an error is an InitCommandError
func IsInitCommandError(err error) bool {
	_, ok := err.(*InitCommandError)
	return ok
}

// NewInitCommandError creates a new InitCommandError
func NewInitCommandError(failedCommands []string) *InitCommandError {
	return &InitCommandError{
		FailedCommands: failedCommands,
	}
}

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

	// Components
	closeDialog components.CloseDialogModel

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
		closeDialog:  components.NewCloseDialogModel(),
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
		// Close dialog takes highest precedence
		if m.closeDialog.IsVisible() {
			dialogResult, cmd := m.closeDialog.Update(msg)
			m.closeDialog = dialogResult

			// Handle dialog result
			if m.closeDialog.IsConfirmed() {
				if selected := m.ticketDetail.SelectedTicket(); selected != nil {
					reason := m.closeDialog.GetReason()
					return m, m.closeTicketWithReason(selected, reason)
				}
				m.closeDialog.Hide()
			} else if m.closeDialog.IsCancelled() {
				m.closeDialog.Hide()
			}

			return m, cmd
		}

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
		m.closeDialog.SetSize(msg.Width, msg.Height)

	case error:
		m.err = msg
		// Hide close dialog if visible to prevent inconsistent state
		if m.closeDialog.IsVisible() {
			m.closeDialog.Hide()
		}
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
				// First validate if this ticket can be closed
				// Check if it's the current ticket before showing dialog
				current, err := m.manager.GetCurrentTicket(context.Background())
				if err != nil {
					m.err = fmt.Errorf("failed to get current ticket: %w", err)
					return m, nil
				}
				if current == nil {
					m.err = fmt.Errorf("no active ticket - use 'ticketflow start <ticket-id>' first")
					return m, nil
				}
				if current.ID != t.ID {
					m.err = fmt.Errorf("can only close the current active ticket (%s). Selected ticket: %s", current.ID, t.ID)
					return m, nil
				}
				
				// Determine if reason is required based on ticket status:
				// - Todo tickets: require reason (being abandoned without starting work)
				// - Doing tickets: optional reason (normal workflow, closing before PR merge)
				// NOTE: This logic could be moved to a ticket.RequiresCloseReason() method
				// in a future refactor to better separate business logic from UI
				requireReason := t.Status() == ticket.StatusTodo

				// Show close dialog with optional reason
				m.closeDialog.Show(requireReason)
				return m, nil // Dialog handles its own blink command
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

	// Add close dialog overlay if visible
	if m.closeDialog.IsVisible() {
		dialogView := m.closeDialog.View()
		// Center the dialog overlay
		dialogWidth := lipgloss.Width(dialogView)
		dialogHeight := lipgloss.Height(dialogView)
		x := (m.width - dialogWidth) / 2
		y := (m.height - dialogHeight) / 2

		// Overlay dialog on content
		dialogOverlay := lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top,
			lipgloss.NewStyle().Margin(y, 0, 0, x).Render(dialogView))

		// Return dialog overlay on top of content
		return dialogOverlay
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
		currentBranch, err := m.git.CurrentBranch(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}

		// Setup branch or worktree
		worktreePath, initErr := m.setupTicketBranchOrWorktree(t)
		if initErr != nil && !IsInitCommandError(initErr) {
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

// closeTicketWithReason closes a ticket with an optional reason
func (m *Model) closeTicketWithReason(t *ticket.Ticket, reason string) tea.Cmd {
	return func() tea.Msg {
		// Create a context with timeout for the close operation
		ctx, cancel := context.WithTimeout(context.Background(), closeOperationTimeout)
		defer cancel()

		// Check for early cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		// Validate ticket can be closed
		if err := m.validateTicketForClose(t); err != nil {
			return err
		}

		// Check for cancellation before workspace check
		if ctx.Err() != nil {
			return fmt.Errorf("operation cancelled before workspace check: %w", ctx.Err())
		}

		// Check workspace state and get worktree info
		worktreePath, isWorktree, err := m.checkWorkspaceForClose(t)
		if err != nil {
			return err
		}

		// Check for cancellation before committing
		if ctx.Err() != nil {
			return fmt.Errorf("operation cancelled before commit: %w", ctx.Err())
		}

		// Move ticket to done status and commit with reason
		if err := m.moveTicketToDoneAndCommitWithContext(ctx, t, reason); err != nil {
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
		updated, reloadErr := m.manager.Get(context.Background(), t.ID)
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
		dirty, err := m.git.HasUncommittedChanges(context.Background())
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
	logger := log.Global().WithTicket(t.ID)
	var worktreePath string

	if m.config.Worktree.Enabled {
		// Check if worktree already exists
		if exists, err := m.git.HasWorktree(context.Background(), t.ID); err != nil {
			logger.WithError(err).Error("failed to check worktree")
			return "", fmt.Errorf("failed to check worktree: %w", err)
		} else if exists {
			worktreePath := worktree.GetPath(context.Background(), m.git, m.config, m.projectRoot, t.ID)
			logger.Debug("worktree already exists", "path", worktreePath)
			return "", fmt.Errorf("worktree for ticket %s already exists at: %s", t.ID, worktreePath)
		}

		// Create worktree
		baseDir := m.config.GetWorktreePath(m.projectRoot)
		worktreePath = filepath.Join(baseDir, t.ID)

		if err := m.git.AddWorktree(context.Background(), worktreePath, t.ID); err != nil {
			logger.WithError(err).Error("failed to create worktree", "path", worktreePath)
			return "", fmt.Errorf("failed to create worktree: %w", err)
		}
		logger.Debug("created worktree", "path", worktreePath)

		// Run init commands if configured
		if err := m.runWorktreeInitCommands(worktreePath); err != nil {
			// Non-fatal: store as warning to display later
			return worktreePath, err
		}
	} else {
		// Original behavior: create and checkout branch
		if err := m.git.CreateBranch(context.Background(), t.ID); err != nil {
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

	// Create context with timeout
	timeout := m.config.GetInitCommandsTimeout()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var failedCommands []string
	for _, cmd := range m.config.Worktree.InitCommands {
		// Parse the command with proper shell parsing
		parts, err := shellwords.Parse(cmd)
		if err != nil {
			failedCommands = append(failedCommands, fmt.Sprintf("%s (failed to parse: %v)", cmd, err))
			continue
		}
		if len(parts) == 0 {
			continue
		}

		execCmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
		execCmd.Dir = worktreePath
		if err := execCmd.Run(); err != nil {
			// Check if error is due to timeout
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				failedCommands = append(failedCommands, fmt.Sprintf("%s (timed out after %v)", cmd, timeout))
			} else {
				failedCommands = append(failedCommands, fmt.Sprintf("%s (%v)", cmd, err))
			}
		}
	}

	if len(failedCommands) > 0 {
		return NewInitCommandError(failedCommands)
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
	if err := m.manager.Update(context.Background(), t); err != nil {
		// Rollback file move
		_ = os.Rename(newPath, oldPath)
		m.rollbackTicketStart(worktreePath, currentBranch)
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	// Git add both old and new paths
	if err := m.git.Add(context.Background(), oldPath, newPath); err != nil {
		return fmt.Errorf("failed to stage ticket move: %w", err)
	}

	// Commit the move
	if err := m.git.Commit(context.Background(), fmt.Sprintf("Start ticket: %s", t.ID)); err != nil {
		return fmt.Errorf("failed to commit ticket move: %w", err)
	}

	// Set current ticket
	if err := m.manager.SetCurrentTicket(context.Background(), t); err != nil {
		return fmt.Errorf("failed to set current ticket: %w", err)
	}

	return nil
}

// rollbackTicketStart rolls back changes made during ticket start
func (m *Model) rollbackTicketStart(worktreePath, currentBranch string) {
	if m.config.Worktree.Enabled && worktreePath != "" {
		_ = m.git.RemoveWorktree(context.Background(), worktreePath)
	} else {
		_ = m.git.Checkout(context.Background(), currentBranch)
	}
}

// validateTicketForClose validates that a ticket can be closed
func (m *Model) validateTicketForClose(t *ticket.Ticket) error {
	// Get current ticket
	current, err := m.manager.GetCurrentTicket(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get current ticket: %w", err)
	}
	if current == nil {
		return fmt.Errorf("no active ticket - use 'ticketflow start <ticket-id>' first")
	}
	if current.ID != t.ID {
		return fmt.Errorf("can only close the current active ticket (%s). Selected ticket: %s", current.ID, t.ID)
	}
	return nil
}

// checkWorkspaceForClose checks workspace state and returns worktree info
func (m *Model) checkWorkspaceForClose(t *ticket.Ticket) (string, bool, error) {
	var worktreePath string
	var isWorktree bool

	if m.config.Worktree.Enabled {
		// Check if a worktree exists for this ticket
		wt, err := m.git.FindWorktreeByBranch(context.Background(), t.ID)
		if err != nil {
			return "", false, fmt.Errorf("failed to find worktree: %w", err)
		}
		if wt != nil {
			isWorktree = true
			worktreePath = wt.Path

			// Check for uncommitted changes in worktree
			wtGit := git.NewWithTimeout(worktreePath, m.config.GetGitTimeout())
			dirty, err := wtGit.HasUncommittedChanges(context.Background())
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
		dirty, err := m.git.HasUncommittedChanges(context.Background())
		if err != nil {
			return "", false, fmt.Errorf("failed to check git status: %w", err)
		}
		if dirty {
			return "", false, fmt.Errorf("uncommitted changes - please commit before closing")
		}

		// Get current branch
		currentBranch, err := m.git.CurrentBranch(context.Background())
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

// moveTicketToDoneAndCommitWithContext moves ticket to done status and commits the change with context support
func (m *Model) moveTicketToDoneAndCommitWithContext(ctx context.Context, t *ticket.Ticket, reason string) error {
	// Check for cancellation at the start
	if ctx.Err() != nil {
		return fmt.Errorf("operation cancelled: %w", ctx.Err())
	}

	// Update ticket status
	if err := m.closeTicketWithStatus(t, reason); err != nil {
		return fmt.Errorf("failed to close ticket %s: %w", t.ID, err)
	}

	// Check for cancellation before file operations
	if ctx.Err() != nil {
		return fmt.Errorf("operation cancelled before file move: %w", ctx.Err())
	}

	// Move file and update ticket
	oldPath := t.Path
	newPath := filepath.Join(m.config.GetDonePath(m.projectRoot), filepath.Base(t.Path))

	if err := m.moveAndUpdateTicket(ctx, t, oldPath, newPath); err != nil {
		return err
	}

	// Check for cancellation before git operations
	if ctx.Err() != nil {
		return fmt.Errorf("operation cancelled before commit: %w", ctx.Err())
	}

	// Commit changes
	if err := m.commitTicketClose(ctx, t, reason, oldPath, newPath); err != nil {
		return err
	}

	// Remove current ticket link
	return m.manager.SetCurrentTicket(ctx, nil)
}

// closeTicketWithStatus closes the ticket with optional reason
func (m *Model) closeTicketWithStatus(t *ticket.Ticket, reason string) error {
	if reason != "" {
		if err := t.CloseWithReason(reason); err != nil {
			return fmt.Errorf("failed to close ticket with reason: %w", err)
		}
	} else {
		if err := t.Close(); err != nil {
			return fmt.Errorf("failed to close ticket: %w", err)
		}
	}
	return nil
}

// moveAndUpdateTicket moves the ticket file and updates its path
func (m *Model) moveAndUpdateTicket(ctx context.Context, t *ticket.Ticket, oldPath, newPath string) error {
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to move ticket to done: %w", err)
	}

	t.Path = newPath
	if err := m.manager.Update(ctx, t); err != nil {
		_ = os.Rename(newPath, oldPath) // Rollback
		return fmt.Errorf("failed to update ticket: %w", err)
	}
	return nil
}

// commitTicketClose stages and commits the ticket closure
func (m *Model) commitTicketClose(ctx context.Context, t *ticket.Ticket, reason, oldPath, newPath string) error {
	if err := m.git.Add(ctx, oldPath, newPath); err != nil {
		return fmt.Errorf("failed to stage ticket move: %w", err)
	}

	commitMsg := fmt.Sprintf("Close ticket: %s", t.ID)
	if reason != "" {
		commitMsg = fmt.Sprintf("Close ticket: %s\n\nReason: %s", t.ID, reason)
	}

	if err := m.git.Commit(ctx, commitMsg); err != nil {
		return fmt.Errorf("failed to commit ticket move: %w", err)
	}
	return nil
}
