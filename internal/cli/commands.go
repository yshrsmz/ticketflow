package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-shellwords"
	"github.com/yshrsmz/ticketflow/internal/config"
	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/log"
	"github.com/yshrsmz/ticketflow/internal/ticket"
	"github.com/yshrsmz/ticketflow/internal/worktree"
)

// Constants for messages
const (
	msgBranchAlreadyMerged = "Branch was already merged"
)

// StartTicketResult contains the result of starting a ticket with worktree details
type StartTicketResult struct {
	// Ticket is the ticket that was started, with updated status and timestamps
	Ticket *ticket.Ticket
	// WorktreePath is the filesystem path to the created worktree (empty if worktrees disabled)
	WorktreePath string
	// ParentBranch is the branch that was checked out from (typically "main")
	ParentBranch string
	// InitCommandsExecuted indicates whether initialization commands were successfully run
	InitCommandsExecuted bool
}

// CleanWorktreesResult represents the result of cleaning worktrees
type CleanWorktreesResult struct {
	// CleanedWorktrees is the list of worktrees that were removed
	CleanedWorktrees []string
	// CleanedCount is the number of worktrees that were cleaned
	CleanedCount int
	// FailedWorktrees is the list of worktrees that failed to be removed
	FailedWorktrees []string
	// TotalWorktrees is the total number of worktrees examined
	TotalWorktrees int
	// ActiveTickets is the number of active tickets in doing status
	ActiveTickets int
}

// App represents the CLI application
type App struct {
	Config       *config.Config
	Git          git.GitClient
	Manager      ticket.TicketManager
	ProjectRoot  string
	workingDir   string        // Working directory for the app (defaults to ".")
	Output       *OutputWriter // Output writer for formatted output
	StatusWriter StatusWriter  // Status writer for progress messages
}

// AppOption represents an option for creating a new App
type AppOption func(*App)

// WithGitClient sets a custom GitClient
func WithGitClient(client git.GitClient) AppOption {
	return func(a *App) {
		a.Git = client
	}
}

// WithTicketManager sets a custom TicketManager
func WithTicketManager(manager ticket.TicketManager) AppOption {
	return func(a *App) {
		a.Manager = manager
	}
}

// WithWorkingDirectory sets a custom working directory
func WithWorkingDirectory(dir string) AppOption {
	return func(a *App) {
		a.workingDir = dir
	}
}

// WithOutputWriter sets a custom output writer
func WithOutputWriter(writer *OutputWriter) AppOption {
	return func(a *App) {
		a.Output = writer
	}
}

// NewAppWithOptions creates a new CLI application with custom options
func NewAppWithOptions(ctx context.Context, opts ...AppOption) (*App, error) {
	// Create app with default working directory
	app := &App{
		workingDir: ".", // Default to current directory
	}

	// Apply options first to allow overriding workingDir
	for _, opt := range opts {
		opt(app)
	}

	// Find project root (with .git directory) using configured working directory
	projectRoot, err := git.FindProjectRoot(ctx, app.workingDir)
	if err != nil {
		return nil, NewError(ErrNotGitRepo, "Not in a git repository", "",
			[]string{
				"Navigate to your project root directory",
				"Initialize a new git repository with 'git init'",
			})
	}

	// Load config with context
	cfg, err := config.LoadWithContext(ctx, projectRoot)
	if err != nil {
		return nil, NewError(ErrConfigNotFound, "Ticket system not initialized", "",
			[]string{
				"Run 'ticketflow init' to initialize",
				"Navigate to the project root directory",
			})
	}

	app.Config = cfg
	app.ProjectRoot = projectRoot

	// Set defaults if not provided
	if app.Git == nil {
		app.Git = git.NewWithTimeout(projectRoot, app.Config.GetGitTimeout())
	}
	if app.Manager == nil {
		app.Manager = ticket.NewManager(cfg, projectRoot)
	}
	if app.Output == nil {
		app.Output = NewOutputWriter(nil, nil, FormatText)
	}
	// Initialize StatusWriter based on output format
	if app.StatusWriter == nil {
		format := FormatText
		if app.Output != nil {
			format = app.Output.GetFormat()
		}
		app.StatusWriter = NewStatusWriter(os.Stdout, format)
	}

	return app, nil
}

// NewApp creates a new CLI application
func NewApp(ctx context.Context) (*App, error) {
	return NewAppWithOptions(ctx)
}

// InitCommand initializes the ticket system (doesn't require existing config)
func InitCommand(ctx context.Context) error {
	return InitCommandWithWorkingDir(ctx, ".")
}

// InitCommandWithWorkingDir initializes the ticket system with a specific working directory
func InitCommandWithWorkingDir(ctx context.Context, workingDir string) error {
	logger := log.Global().WithOperation("init")

	projectRoot, err := git.FindProjectRoot(ctx, workingDir)
	if err != nil {
		logger.WithError(err).Error("not in a git repository")
		return NewError(ErrNotGitRepo, "Not in a git repository", "", nil)
	}
	logger.Debug("found project root", "path", projectRoot)

	// Create default config
	cfg := config.Default()
	configPath := filepath.Join(projectRoot, ".ticketflow.yaml")

	// Check if already exists
	if _, err := os.Stat(configPath); err == nil {
		logger.Info("ticket system already initialized")
		status := NewStatusWriter(os.Stdout, FormatText)
		status.Println("Ticket system already initialized")
		return nil
	}

	// Save config with context
	if err := cfg.SaveWithContext(ctx, configPath); err != nil {
		logger.WithError(err).Error("failed to save config")
		return fmt.Errorf("failed to save config: %w", err)
	}
	logger.Info("saved configuration", "path", configPath)

	// Create directory structure
	ticketsDir := filepath.Join(projectRoot, cfg.Tickets.Dir)
	todoDir := filepath.Join(ticketsDir, cfg.Tickets.TodoDir)
	doingDir := filepath.Join(ticketsDir, cfg.Tickets.DoingDir)
	doneDir := filepath.Join(ticketsDir, cfg.Tickets.DoneDir)

	// Create all directories
	for _, dir := range []string{ticketsDir, todoDir, doingDir, doneDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.WithError(err).Error("failed to create directory", "dir", dir)
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		logger.Debug("created directory", "dir", dir)
	}

	// Update .gitignore
	gitignorePath := filepath.Join(projectRoot, GitignoreFile)
	if err := updateGitignore(gitignorePath); err != nil {
		logger.WithError(err).Error("failed to update .gitignore")
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}
	logger.Info("updated .gitignore")

	logger.Info("ticket system initialized successfully")
	status := NewStatusWriter(os.Stdout, FormatText)
	status.Println("Initialized ticket system successfully")
	status.Printf("Configuration saved to: %s\n", configPath)
	status.Printf("Tickets directory: %s\n", ticketsDir)

	return nil
}

// validateSlug checks if the slug is valid
func (app *App) validateSlug(slug string) error {
	if !ticket.IsValidSlug(slug) {
		logger := log.Global().WithOperation("new_ticket")
		logger.Error("invalid slug format", slog.String("slug", slug))
		return NewError(ErrTicketInvalid, "Invalid slug format",
			fmt.Sprintf("Slug '%s' contains invalid characters", slug),
			[]string{
				"Use only lowercase letters (a-z)",
				"Use only numbers (0-9)",
				"Use only hyphens (-) for separation",
			})
	}
	return nil
}

// warnIfParentDone warns if the parent ticket is already done
func (app *App) warnIfParentDone(parentTicket *ticket.Ticket, parentID string) {
	if parentTicket.Status() == ticket.StatusDone {
		app.Output.Printf("⚠️  Warning: Parent ticket '%s' is already done\n", parentID)
	}
}

// checkCircularDependency checks if adding newTicketID as a child of parentID would create a circular dependency
func (app *App) checkCircularDependency(ctx context.Context, parentID, newTicketID string) error {
	// Check if the parent ticket has the new ticket as an ancestor
	currentID := parentID
	visited := make(map[string]bool)

	for currentID != "" {
		// Prevent infinite loops in case of existing circular dependencies
		if visited[currentID] {
			break
		}
		visited[currentID] = true

		// Check if we've reached the new ticket ID
		if currentID == newTicketID {
			return NewError(ErrTicketInvalid, "Circular dependency detected",
				fmt.Sprintf("Creating this relationship would form a circular dependency: %s → %s", newTicketID, parentID),
				[]string{
					"Choose a different parent ticket",
					"Check the ticket hierarchy with 'ticketflow show'",
				})
		}

		// Get the current ticket to check its parent
		currentTicket, err := app.Manager.Get(ctx, currentID)
		if err != nil {
			// If we can't get the ticket, assume no circular dependency
			break
		}

		// Extract parent from related field
		currentID = app.extractParentTicketID(currentTicket)
	}

	return nil
}

// validateExplicitParent validates an explicitly provided parent ticket
func (app *App) validateExplicitParent(ctx context.Context, explicitParent, slug string) (string, error) {
	logger := log.Global().WithOperation("new_ticket")

	// Prevent self-parenting (check this first before checking if parent exists)
	// Get the generated ticket ID for comparison
	generatedID := ticket.GenerateID(slug)
	if explicitParent == slug || explicitParent == generatedID {
		return "", NewError(ErrTicketInvalid, "Invalid parent relationship",
			"A ticket cannot be its own parent",
			[]string{
				"Choose a different parent ticket",
				"Or create a top-level ticket without --parent",
			})
	}

	// Validate that the parent ticket exists
	parentTicket, err := app.Manager.Get(ctx, explicitParent)
	if err != nil {
		logger.Error("parent ticket not found", slog.String("parent", explicitParent))
		return "", NewError(ErrTicketNotFound, "Parent ticket not found",
			fmt.Sprintf("Ticket '%s' does not exist", explicitParent),
			[]string{
				"Check the ticket ID is correct",
				"Use 'ticketflow list' to see available tickets",
			})
	}

	// Check for circular dependencies
	if err := app.checkCircularDependency(ctx, explicitParent, generatedID); err != nil {
		return "", err
	}

	// Warn if parent ticket is done (but still allow it)
	app.warnIfParentDone(parentTicket, explicitParent)

	app.Output.Printf("Creating sub-ticket with parent: %s\n", explicitParent)
	return explicitParent, nil
}

// detectImplicitParent detects parent ticket from current branch
func (app *App) detectImplicitParent(ctx context.Context) (string, error) {
	currentBranch, err := app.Git.CurrentBranch(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	if currentBranch != app.Config.Git.DefaultBranch {
		// Check if current branch is a ticket
		if parentTicket, err := app.Manager.Get(ctx, currentBranch); err == nil {
			app.Output.Printf("Creating ticket in branch: %s\n", currentBranch)
			// Warn if parent ticket is done (but still allow it)
			app.warnIfParentDone(parentTicket, currentBranch)
			return currentBranch, nil
		}
	}

	return "", nil
}

// resolveParentTicket resolves the parent ticket from explicit or implicit sources
func (app *App) resolveParentTicket(ctx context.Context, explicitParent, slug string) (string, error) {
	if explicitParent != "" {
		return app.validateExplicitParent(ctx, explicitParent, slug)
	}
	return app.detectImplicitParent(ctx)
}

// outputTicketCreated outputs the result of ticket creation
// NewTicket creates a new ticket
func (app *App) NewTicket(ctx context.Context, slug string, explicitParent string) (*ticket.Ticket, error) {
	logger := log.Global().WithOperation("new_ticket")

	// Validate slug
	if err := app.validateSlug(slug); err != nil {
		return nil, err
	}

	logger.Debug("creating new ticket", slog.String("slug", slug), slog.String("explicit_parent", explicitParent))

	// Resolve parent ticket
	parentTicketID, err := app.resolveParentTicket(ctx, explicitParent, slug)
	if err != nil {
		return nil, err
	}

	// Create ticket
	t, err := app.Manager.Create(ctx, slug)
	if err != nil {
		logger.WithError(err).Error("failed to create ticket", slog.String("slug", slug))
		return nil, ConvertError(err)
	}
	logger.Info("created ticket", "ticket_id", t.ID, "path", t.Path)

	// If this is a sub-ticket, update its metadata
	if parentTicketID != "" {
		logger.Debug("creating sub-ticket", "parent", parentTicketID)
		// Add parent relationship
		t.Related = append(t.Related, fmt.Sprintf("parent:%s", parentTicketID))
		if err := app.Manager.Update(ctx, t); err != nil {
			logger.WithError(err).Error("failed to update ticket metadata", slog.String("ticket_id", t.ID), slog.String("parent", parentTicketID))
			return nil, fmt.Errorf("failed to update ticket metadata: %w", err)
		}
		logger.Info("created sub-ticket", "ticket_id", t.ID, "parent", parentTicketID)
	}

	// Return ticket (output is handled by command layer)
	return t, nil
}

// ListTickets lists tickets
func (app *App) ListTickets(ctx context.Context, status ticket.Status, count int, format OutputFormat) error {
	// Convert Status to StatusFilter
	var statusFilter ticket.StatusFilter
	switch status {
	case "":
		statusFilter = ticket.StatusFilterActive
	case ticket.StatusTodo:
		statusFilter = ticket.StatusFilterTodo
	case ticket.StatusDoing:
		statusFilter = ticket.StatusFilterDoing
	case ticket.StatusDone:
		statusFilter = ticket.StatusFilterDone
	case StatusAll:
		statusFilter = ticket.StatusFilterAll
	default:
		return fmt.Errorf("invalid status filter: %s", status)
	}

	tickets, err := app.Manager.List(ctx, statusFilter)
	if err != nil {
		return err
	}

	// Limit count
	if count > 0 && len(tickets) > count {
		tickets = tickets[:count]
	}

	// Calculate summary counts
	allTickets, err := app.Manager.List(ctx, ticket.StatusFilterAll)
	if err != nil {
		return err
	}

	todoCount, doingCount, doneCount := app.countTicketsByStatus(allTickets)

	// Create TicketListResult
	result := &TicketListResult{
		Tickets: tickets,
		Count: map[string]int{
			"total": len(allTickets),
			"todo":  todoCount,
			"doing": doingCount,
			"done":  doneCount,
		},
	}

	return app.Output.PrintResult(result)
}

// StartTicket starts working on a ticket
func (app *App) StartTicket(ctx context.Context, ticketID string, force bool) (*StartTicketResult, error) {
	logger := log.Global().WithOperation("start_ticket").WithTicket(ticketID)
	logger.Info("starting ticket")

	// Get and validate the ticket
	t, err := app.validateTicketForStart(ctx, ticketID, force)
	if err != nil {
		logger.WithError(err).Error("failed to validate ticket")
		return nil, err
	}

	// Check workspace state
	if err := app.checkWorkspaceForStart(ctx); err != nil {
		return nil, err
	}

	// Get current branch and detect parent
	currentBranch, parentBranch, err := app.detectParentBranch(ctx)
	if err != nil {
		return nil, err
	}

	// Setup branch for the ticket
	if err := app.setupTicketBranch(ctx, t, currentBranch); err != nil {
		return nil, err
	}

	// Check if worktree already exists (for worktree mode)
	if err := app.checkExistingWorktree(ctx, t, force); err != nil {
		return nil, err
	}

	// Update parent relationship if needed
	if err := app.updateParentRelationship(ctx, t, parentBranch, currentBranch); err != nil {
		return nil, err
	}

	// Move ticket to doing status (skip if already in doing and using force)
	if t.Status() != ticket.StatusDoing {
		if err := app.moveTicketToDoing(ctx, t, currentBranch); err != nil {
			return nil, err
		}
	}

	// Now create worktree AFTER committing (for worktree mode)
	worktreePath, err := app.createAndSetupWorktree(ctx, t)
	if err != nil {
		return nil, err
	}

	// Check if init commands were executed
	initCommandsExecuted := len(app.Config.Worktree.InitCommands) > 0 && worktreePath != ""

	// Print success message for text format
	app.printStartSuccessMessage(t, worktreePath, parentBranch)

	return &StartTicketResult{
		Ticket:               t,
		WorktreePath:         worktreePath,
		ParentBranch:         parentBranch,
		InitCommandsExecuted: initCommandsExecuted,
	}, nil
}

// closeCurrentTicketInternal handles the common logic for closing the current ticket
func (app *App) closeCurrentTicketInternal(ctx context.Context, reason string, force bool) (*ticket.Ticket, error) {
	operation := "close_ticket"
	if reason != "" {
		operation = "close_ticket_with_reason"
	}
	logger := log.Global().WithOperation(operation)

	// Validate current ticket for close
	current, worktreePath, err := app.validateTicketForClose(ctx, force)
	if err != nil {
		logger.WithError(err).Error("failed to validate ticket for close")
		return nil, err
	}

	logger = logger.WithTicket(current.ID)

	// Update ticket status
	if reason != "" {
		logger.Info("closing ticket with reason", "reason", reason)
		if err := current.CloseWithReason(reason); err != nil {
			return nil, fmt.Errorf("failed to close ticket with reason: %w", err)
		}
	} else {
		logger.Info("closing ticket")
		if err := current.Close(); err != nil {
			return nil, err
		}
	}

	// Move ticket to done status (this also commits with the reason)
	// Pass true because closeCurrentTicketInternal always operates on the current ticket
	if err := app.moveTicketToDoneWithReason(ctx, current, reason, true); err != nil {
		return nil, err
	}

	// Calculate work duration
	duration := app.calculateWorkDuration(current)

	// Extract parent ticket ID
	parentTicketID := app.extractParentTicketID(current)

	// Print success message with next steps
	app.printCloseSuccessMessage(current, duration, parentTicketID, worktreePath)

	if reason != "" {
		logger.Info("ticket closed successfully with reason", "duration", duration, "reason", reason)
	} else {
		logger.Info("ticket closed successfully", "duration", duration)
	}
	return current, nil
}

// CloseTicket closes the current ticket
func (app *App) CloseTicket(ctx context.Context, force bool) (*ticket.Ticket, error) {
	return app.closeCurrentTicketInternal(ctx, "", force)
}

// CloseTicketWithReason closes the current ticket with a reason
func (app *App) CloseTicketWithReason(ctx context.Context, reason string, force bool) (*ticket.Ticket, error) {
	// Validate that reason is not empty or just whitespace
	if strings.TrimSpace(reason) == "" {
		return nil, NewError(ErrValidation, "Empty reason",
			"Reason cannot be empty or just whitespace",
			[]string{"Provide a meaningful reason for closing the ticket"})
	}
	return app.closeCurrentTicketInternal(ctx, reason, force)
}

// checkBranchMerged checks if a branch has been merged to the default branch
func (app *App) checkBranchMerged(ctx context.Context, ticketID string) (bool, error) {
	if app.Config.Git.DefaultBranch == "" {
		return false, nil
	}
	return app.Git.IsBranchMerged(ctx, ticketID, app.Config.Git.DefaultBranch)
}

// validateTicketByID validates that a ticket can be closed by ID
func (app *App) validateTicketByID(ctx context.Context, ticketID string) (*ticket.Ticket, error) {
	ticket, err := app.Manager.Get(ctx, ticketID)
	if err != nil {
		return nil, NewError(ErrTicketNotFound, "Ticket not found",
			fmt.Sprintf("Ticket '%s' does not exist", ticketID),
			[]string{"Check ticket ID and try again", "List available tickets: ticketflow list"})
	}

	if ticket.ClosedAt.Time != nil {
		return nil, NewError(ErrTicketAlreadyClosed, "Ticket already closed",
			fmt.Sprintf("Ticket '%s' is already closed", ticketID),
			nil)
	}

	return ticket, nil
}

// closeAndCommitTicket closes the ticket, saves it, and commits the changes
func (app *App) closeAndCommitTicket(ctx context.Context, ticket *ticket.Ticket, reason string, branchMerged bool) error {
	logger := log.Global().WithOperation("close_and_commit").WithTicket(ticket.ID)

	// Close the ticket
	if reason != "" {
		logger.Info("closing ticket with reason", "reason", reason)
		if err := ticket.CloseWithReason(reason); err != nil {
			return fmt.Errorf("failed to close ticket with reason: %w", err)
		}
	} else if branchMerged {
		logger.Info("closing ticket", "note", msgBranchAlreadyMerged)
		// Branch is merged, close normally without implicit reason
		if err := ticket.Close(); err != nil {
			return fmt.Errorf("failed to close merged ticket: %w", err)
		}
	} else {
		// Should not reach here as validation should prevent this
		return errors.New("cannot close ticket without reason when branch is not merged")
	}

	// Move ticket to done status (this also saves the ticket and commits with the reason)
	// Note: moveTicketToDoneWithReason will call Update, so we don't need to do it here
	// Pass false because CloseTicketByID may close any ticket, not necessarily the current one
	if err := app.moveTicketToDoneWithReason(ctx, ticket, reason, false); err != nil {
		return fmt.Errorf("failed to move ticket to done: %w", err)
	}

	return nil
}

// printCloseByIDSuccessMessage prints the success message after closing a ticket by ID
func (app *App) printCloseByIDSuccessMessage(ticket *ticket.Ticket, reason string, branchMerged bool, worktreePath string) {
	duration := app.calculateWorkDuration(ticket)

	app.Output.Printf("\n✅ Ticket closed: %s\n", ticket.ID)
	app.Output.Printf("   Description: %s\n", ticket.Description)
	if reason != "" {
		app.Output.Printf("   Reason: %s\n", reason)
	} else if branchMerged {
		app.Output.Printf("   Note: %s\n", msgBranchAlreadyMerged)
	}
	if duration != "" {
		app.Output.Printf("   Duration: %s\n", duration)
	}
	app.Output.Printf("   Committed: \"Close ticket: %s\"\n", ticket.ID)

	// Suggest cleanup if worktree exists
	if worktreePath != "" {
		app.Output.Printf("\n💡 Ticket has a worktree. Run this to clean up:\n")
		app.Output.Printf("   ticketflow cleanup %s\n", ticket.ID)
	}
}

// CloseTicketByID closes a specific ticket by ID
func (app *App) CloseTicketByID(ctx context.Context, ticketID, reason string, force bool) (*ticket.Ticket, error) {
	logger := log.Global().WithOperation("close_ticket_by_id").WithTicket(ticketID)

	// First check if this is the current ticket
	current, _ := app.Manager.GetCurrentTicket(ctx)
	if current != nil && current.ID == ticketID {
		// This is the current ticket, use normal close logic
		if reason != "" {
			return app.CloseTicketWithReason(ctx, reason, force)
		}
		return app.CloseTicket(ctx, force)
	}

	// Validate the ticket
	ticket, err := app.validateTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Check if branch is merged
	branchMerged := false
	merged, err := app.checkBranchMerged(ctx, ticketID)
	if err != nil {
		logger.WithError(err).Warn("failed to check if branch is merged, assuming not merged")
	} else {
		branchMerged = merged
	}

	// If no reason provided and branch not merged, require reason
	if reason == "" && !branchMerged {
		return nil, NewError(ErrValidation, "Reason required",
			"Closing a ticket without being in its worktree requires a reason",
			[]string{
				fmt.Sprintf("Provide a reason: ticketflow close %s --reason \"explanation\"", ticketID),
				"Or switch to the ticket's worktree to close normally",
			})
	}

	// Check for worktree
	var worktreePath string
	if app.Config.Worktree.Enabled {
		wt, err := app.Git.FindWorktreeByBranch(ctx, ticketID)
		if err == nil && wt != nil {
			worktreePath = wt.Path
		}
	}

	// Close and commit the ticket
	if err := app.closeAndCommitTicket(ctx, ticket, reason, branchMerged); err != nil {
		return nil, err
	}

	// Print success message
	app.printCloseByIDSuccessMessage(ticket, reason, branchMerged, worktreePath)

	logger.Info("ticket closed successfully", "duration", app.calculateWorkDuration(ticket), "reason", reason, "branchMerged", branchMerged)
	return ticket, nil
}

// RestoreCurrentTicket restores the current ticket symlink
func (app *App) RestoreCurrentTicket(ctx context.Context) (*ticket.Ticket, error) {
	// Get current branch
	branch, err := app.Git.CurrentBranch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Try to get ticket by branch name
	t, err := app.Manager.Get(ctx, branch)
	if err != nil {
		return nil, ConvertError(fmt.Errorf("no ticket found for branch %s", branch))
	}

	// Set current ticket
	if err := app.Manager.SetCurrentTicket(ctx, t); err != nil {
		return nil, fmt.Errorf("failed to set current ticket: %w", err)
	}

	app.Output.Printf("Restored current ticket link: %s\n", t.ID)
	return t, nil
}

// Status shows the current status
func (app *App) Status(ctx context.Context, format OutputFormat) error {
	// Get current ticket
	current, err := app.Manager.GetCurrentTicket(ctx)
	if err != nil {
		return ConvertError(err)
	}

	// Get current branch
	branch, err := app.Git.CurrentBranch(ctx)
	if err != nil {
		return err
	}

	// Get ticket stats
	allTickets, err := app.Manager.List(ctx, ticket.StatusFilterAll)
	if err != nil {
		return err
	}

	todoCount, doingCount, doneCount := app.countTicketsByStatus(allTickets)

	if format == FormatJSON {
		return app.formatStatusJSON(branch, current, allTickets, todoCount, doingCount, doneCount)
	}

	// Text format
	app.printStatusText(ctx, branch, current, allTickets, todoCount, doingCount, doneCount)

	return nil
}

// Helper functions

func updateGitignore(path string) error {
	// Read existing .gitignore
	content := ""
	if data, err := os.ReadFile(path); err == nil {
		content = string(data)
	}

	// Check if already contains our entries
	if strings.Contains(content, ticket.CurrentTicketFile) {
		return nil
	}

	// Append our entries
	toAdd := fmt.Sprintf("\n# TicketFlow\n%s\n%s\n", ticket.CurrentTicketFile, WorktreesDir)

	// Write back
	return os.WriteFile(path, []byte(content+toAdd), ticket.DefaultPermission)
}

func (app *App) outputTicketListText(tickets []*ticket.Ticket) error {
	if len(tickets) == 0 {
		app.Output.Println("No tickets found")
		return nil
	}

	// Find max ID length for alignment
	maxIDLen := 0
	for _, t := range tickets {
		if len(t.ID) > maxIDLen {
			maxIDLen = len(t.ID)
		}
	}

	// Print header
	app.Output.Printf("%-*s  %-6s  %-3s  %s\n", maxIDLen, "ID", "STATUS", "PRI", "DESCRIPTION")
	app.Output.Println(strings.Repeat("-", maxIDLen+50))

	// Print tickets
	for _, t := range tickets {
		status := string(t.Status())
		priority := fmt.Sprintf("%d", t.Priority)

		// Truncate description if too long
		desc := t.Description
		maxDescLen := 50
		if len(desc) > maxDescLen {
			desc = desc[:maxDescLen-3] + "..."
		}

		app.Output.Printf("%-*s  %-6s  %-3s  %s\n", maxIDLen, t.ID, status, priority, desc)
	}

	return nil
}

func (app *App) outputTicketListJSON(ctx context.Context, tickets []*ticket.Ticket) error {
	ticketList := make([]map[string]interface{}, len(tickets))
	for i, t := range tickets {
		// Get worktree path if exists
		var worktreePath string
		if app.Config.Worktree.Enabled && t.HasWorktree() {
			wt, err := app.Git.FindWorktreeByBranch(ctx, t.ID)
			if err == nil && wt != nil {
				worktreePath = wt.Path
			}
		}
		ticketList[i] = ticketToJSON(t, worktreePath)
	}

	// Always calculate full summary from all tickets
	allTickets, err := app.Manager.List(ctx, ticket.StatusFilterAll)
	if err != nil {
		return err
	}

	todoCount, doingCount, doneCount := app.countTicketsByStatus(allTickets)

	output := map[string]interface{}{
		"tickets": ticketList,
		"summary": map[string]int{
			"total": len(allTickets),
			"todo":  todoCount,
			"doing": doingCount,
			"done":  doneCount,
		},
	}

	return app.Output.PrintJSON(output)
}

// ListWorktrees lists all worktrees
func (app *App) ListWorktrees(ctx context.Context, format OutputFormat) error {
	worktrees, err := app.Git.ListWorktrees(ctx)
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Create WorktreeListResult
	result := &WorktreeListResult{
		Worktrees: worktrees,
	}

	return app.Output.PrintResult(result)
}

// CleanWorktrees removes orphaned worktrees
func (app *App) CleanWorktrees(ctx context.Context) (*CleanWorktreesResult, error) {
	logger := log.Global()

	result := &CleanWorktreesResult{
		CleanedWorktrees: []string{},
		FailedWorktrees:  []string{},
	}

	// First prune to clean up git's internal state
	if err := app.Git.PruneWorktrees(ctx); err != nil {
		return nil, fmt.Errorf("failed to prune worktrees: %w", err)
	}

	// Get all worktrees
	worktrees, err := app.Git.ListWorktrees(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Get all active tickets
	activeTickets, err := app.Manager.List(ctx, ticket.StatusFilterDoing)
	if err != nil {
		return nil, fmt.Errorf("failed to list active tickets: %w", err)
	}

	result.ActiveTickets = len(activeTickets)

	// Create a map of active ticket IDs
	activeMap := make(map[string]bool)
	for _, t := range activeTickets {
		activeMap[t.ID] = true
	}

	// Count total worktrees (excluding main)
	for _, wt := range worktrees {
		if wt.Branch != "" && wt.Branch != app.Config.Git.DefaultBranch {
			result.TotalWorktrees++
		}
	}

	// Find and remove orphaned worktrees
	for _, wt := range worktrees {
		// Skip main worktree
		if wt.Branch == "" || wt.Branch == app.Config.Git.DefaultBranch {
			continue
		}

		// Check if this worktree corresponds to an active ticket
		if !activeMap[wt.Branch] {
			app.Output.Printf("Removing orphaned worktree: %s (branch: %s)\n", wt.Path, wt.Branch)
			if err := app.Git.RemoveWorktree(ctx, wt.Path); err != nil {
				logger.WithError(err).Warn("failed to remove worktree", "path", wt.Path)
				app.Output.Printf("Warning: Failed to remove worktree: %v\n", err)
				result.FailedWorktrees = append(result.FailedWorktrees, wt.Branch)
			} else {
				result.CleanedWorktrees = append(result.CleanedWorktrees, wt.Branch)
				result.CleanedCount++
			}
		}
	}

	if result.CleanedCount == 0 {
		app.Output.Println("No orphaned worktrees found")
	} else {
		app.Output.Printf("Cleaned %d orphaned worktree(s)\n", result.CleanedCount)
	}

	return result, nil
}

// CleanupTicket cleans up a ticket after PR merge
func (app *App) CleanupTicket(ctx context.Context, ticketID string, force bool) (*ticket.Ticket, error) {
	// Get the ticket to verify it exists and is done
	t, err := app.Manager.Get(ctx, ticketID)
	if err != nil {
		return nil, ConvertError(err)
	}

	// Check if ticket is done
	if t.Status() != ticket.StatusDone {
		return nil, NewError(ErrTicketNotDone, "Ticket is not done",
			fmt.Sprintf("Ticket %s is in '%s' status, not 'done'", t.ID, t.Status()),
			[]string{
				"Close the ticket first: ticketflow close",
				"Or manually move the ticket to done directory",
			})
	}

	// Get current branch to restore later
	currentBranch, err := app.Git.CurrentBranch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Make sure we're on the default branch
	if currentBranch != app.Config.Git.DefaultBranch {
		if err := app.Git.Checkout(ctx, app.Config.Git.DefaultBranch); err != nil {
			return nil, fmt.Errorf("failed to checkout default branch: %w", err)
		}
	}

	// Check if worktree exists
	wt, err := app.Git.FindWorktreeByBranch(ctx, t.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find worktree: %w", err)
	}

	// Show what will be done
	app.Output.Printf("\n🗑️  Cleanup for ticket: %s\n", t.ID)
	app.Output.Printf("   Description: %s\n", t.Description)
	app.Output.Printf("\nThis will:\n")
	if wt != nil {
		app.Output.Printf("  • Remove worktree: %s\n", wt.Path)
	}
	app.Output.Printf("  • Delete local branch: %s\n", t.ID)

	// Confirmation prompt if not forced
	if !force {
		app.Output.Printf("\nAre you sure? (y/N): ")

		var response string
		// TODO: Consider using a configurable input reader for better testability
		_, _ = fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			app.Output.Println("\n❌ Cleanup cancelled")
			return nil, nil
		}
	}

	app.Output.Printf("\n🔧 Performing cleanup...\n")

	// Remove worktree if it exists
	if wt != nil {
		app.Output.Printf("🌳 Removing worktree: %s\n", wt.Path)
		if err := app.Git.RemoveWorktree(ctx, wt.Path); err != nil {
			return nil, fmt.Errorf("failed to remove worktree at %s for ticket %s: %w", wt.Path, ticketID, err)
		}
	}

	// Delete local branch
	app.Output.Printf("🌿 Deleting local branch: %s\n", t.ID)
	if _, err := app.Git.Exec(ctx, "branch", "-D", t.ID); err != nil {
		// Branch might not exist locally, which is fine
		app.Output.Printf("⚠️  Note: Local branch %s not found or already deleted\n", t.ID)
	}

	app.Output.Printf("\n✅ Cleanup completed successfully!\n")
	app.Output.Printf("\n📋 What's next:\n")
	app.Output.Printf("• Start a new ticket: ticketflow new <slug>\n")
	app.Output.Printf("• View open tickets: ticketflow list --status todo\n")
	app.Output.Printf("• Check active work: ticketflow list --status doing\n")
	return t, nil
}

// validateTicketForStart validates that a ticket can be started
func (app *App) validateTicketForStart(ctx context.Context, ticketID string, force bool) (*ticket.Ticket, error) {
	// Get the ticket
	t, err := app.Manager.Get(ctx, ticketID)
	if err != nil {
		return nil, ConvertError(err)
	}

	// Check if already started
	if t.Status() == ticket.StatusDoing {
		// If force is enabled and worktrees are used, allow restarting
		if force && app.Config.Worktree.Enabled {
			return t, nil
		}
		return nil, NewError(ErrTicketAlreadyStarted, "Ticket already started",
			fmt.Sprintf("Ticket %s is already in progress", t.ID), nil)
	}

	return t, nil
}

// checkWorkspaceForStart checks if the workspace is ready to start a ticket
func (app *App) checkWorkspaceForStart(ctx context.Context) error {
	// Check for uncommitted changes (only if not using worktrees)
	if !app.Config.Worktree.Enabled {
		dirty, err := app.Git.HasUncommittedChanges(ctx)
		if err != nil {
			return fmt.Errorf("failed to check git status: %w", err)
		}
		if dirty {
			return NewError(ErrGitDirtyWorkspace, "Uncommitted changes detected",
				"Please commit or stash your changes before starting a ticket",
				[]string{
					"Commit your changes: git commit -am 'Your message'",
					"Stash your changes: git stash",
				})
		}
	}
	return nil
}

// detectParentBranch detects if we're starting from a parent ticket branch
func (app *App) detectParentBranch(ctx context.Context) (currentBranch string, parentBranch string, err error) {
	currentBranch, err = app.Git.CurrentBranch(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get current branch: %w", err)
	}

	// Check if we're starting from a worktree (for sub-tickets)
	if currentBranch != app.Config.Git.DefaultBranch {
		// Verify this is a valid ticket branch
		if _, err := app.Manager.Get(ctx, currentBranch); err == nil {
			// This is a sub-ticket being created from a parent ticket
			parentBranch = currentBranch
		} else {
			return "", "", NewError(ErrTicketInvalid, "Invalid branch for starting ticket",
				fmt.Sprintf("Currently on branch '%s', which is not a ticket branch", currentBranch),
				[]string{
					fmt.Sprintf("Switch to default branch: git checkout %s", app.Config.Git.DefaultBranch),
					"Or start from a valid ticket branch",
				})
		}
	}

	return currentBranch, parentBranch, nil
}

// setupTicketBranch creates and sets up the branch for the ticket
func (app *App) setupTicketBranch(ctx context.Context, t *ticket.Ticket, currentBranch string) error {
	// For non-worktree mode, create and checkout branch immediately
	if !app.Config.Worktree.Enabled {
		if err := app.Git.CreateBranch(ctx, t.ID); err != nil {
			return fmt.Errorf("failed to create branch %s: %w", t.ID, err)
		}
	}
	return nil
}

// updateParentRelationship updates the parent relationship if needed
func (app *App) updateParentRelationship(ctx context.Context, t *ticket.Ticket, parentBranch string, currentBranch string) error {
	if parentBranch != "" && parentBranch != currentBranch {
		// Add parent to Related field
		parentRef := fmt.Sprintf("parent:%s", parentBranch)
		hasParent := false
		for _, rel := range t.Related {
			if rel == parentRef {
				hasParent = true
				break
			}
		}
		if !hasParent {
			t.Related = append(t.Related, parentRef)
		}
		if err := app.Manager.Update(ctx, t); err != nil {
			return fmt.Errorf("failed to update parent relationship: %w", err)
		}
	}
	return nil
}

// moveTicketToDoing moves the ticket to doing status and commits the change
func (app *App) moveTicketToDoing(ctx context.Context, t *ticket.Ticket, currentBranch string) error {
	// Mark ticket as started
	if err := t.Start(); err != nil {
		return fmt.Errorf("failed to start ticket: %w", err)
	}

	// Move ticket to doing
	doingPath := app.Config.GetDoingPath(app.ProjectRoot)
	newPath := filepath.Join(doingPath, filepath.Base(t.Path))

	// Ensure doing directory exists
	if err := os.MkdirAll(doingPath, 0755); err != nil {
		return fmt.Errorf("failed to create doing directory: %w", err)
	}

	// Store the old path before moving
	oldPath := t.Path

	// Move the file
	if err := os.Rename(t.Path, newPath); err != nil {
		return fmt.Errorf("failed to move ticket to doing: %w", err)
	}

	// Update ticket path
	t.Path = newPath

	// Update the ticket with new path and started timestamp
	if err := app.Manager.Update(ctx, t); err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	// Stage and commit the move - use -A to handle the rename properly
	if err := app.Git.Add(ctx, "-A", filepath.Dir(oldPath), filepath.Dir(newPath)); err != nil {
		return fmt.Errorf("failed to stage ticket move: %w", err)
	}

	if err := app.Git.Commit(ctx, fmt.Sprintf("Start ticket: %s", t.ID)); err != nil {
		return fmt.Errorf("failed to commit ticket move: %w", err)
	}

	// Set as current ticket only in non-worktree mode
	// In worktree mode, the symlink will be created in the worktree by createWorktreeTicketSymlink
	// This prevents duplicate symlinks in both main repo and worktree
	if !app.Config.Worktree.Enabled {
		if err := app.Manager.SetCurrentTicket(ctx, t); err != nil {
			return fmt.Errorf("failed to set current ticket: %w", err)
		}
	}

	// In worktree mode, switch back to original branch
	// In non-worktree mode, stay on the ticket branch to work on it
	if app.Config.Worktree.Enabled && currentBranch != t.ID {
		if err := app.Git.Checkout(ctx, currentBranch); err != nil {
			return fmt.Errorf("failed to switch back to original branch: %w", err)
		}
	}

	return nil
}

// validateTicketForClose validates that a ticket can be closed
func (app *App) validateTicketForClose(ctx context.Context, force bool) (*ticket.Ticket, string, error) {
	// Get current ticket
	current, err := app.Manager.GetCurrentTicket(ctx)
	if err != nil {
		// Check if this is a symlink/readlink error that could be fixed with restore
		// This typically happens when:
		// 1. The current-ticket.md symlink exists but points to a non-existent file
		// 2. The symlink is corrupted or has permission issues
		// 3. The user is in a worktree but the symlink wasn't properly restored
		//
		// GetCurrentTicket wraps os.Readlink errors with "failed to read current ticket link: %w"
		// We rely on this error wrapping to detect readlink failures through the PathError chain
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) && pathErr.Op == "readlink" {
			return nil, "", NewError(ErrTicketNotStarted, "Failed to read current ticket",
				err.Error(),
				[]string{
					"Try restoring the current ticket link: ticketflow restore",
					"Or start a ticket manually: ticketflow start <ticket-id>",
					"List available tickets: ticketflow list",
				})
		}
		return nil, "", ConvertError(err)
	}
	if current == nil {
		return nil, "", NewError(ErrTicketNotStarted, "No active ticket",
			"There is no ticket currently being worked on",
			[]string{
				"Start a ticket first: ticketflow start <ticket-id>",
				"Restore current ticket link if in a worktree: ticketflow restore",
				"List available tickets: ticketflow list",
			})
	}

	var worktreePath string
	var isWorktree bool

	if app.Config.Worktree.Enabled {
		// Check if a worktree exists for this ticket
		wt, err := app.Git.FindWorktreeByBranch(ctx, current.ID)
		if err != nil {
			return nil, "", fmt.Errorf("failed to find worktree: %w", err)
		}
		if wt != nil {
			isWorktree = true
			worktreePath = wt.Path

			// Check for uncommitted changes in worktree
			if !force {
				wtGit := git.NewWithTimeout(worktreePath, app.Config.GetGitTimeout())
				dirty, err := wtGit.HasUncommittedChanges(ctx)
				if err != nil {
					return nil, "", fmt.Errorf("failed to check worktree status: %w", err)
				}
				if dirty {
					return nil, "", NewError(ErrGitDirtyWorkspace, "Uncommitted changes in worktree",
						fmt.Sprintf("Please commit your changes in %s before closing the ticket", worktreePath),
						[]string{
							fmt.Sprintf("cd %s && git commit -am 'Your message'", worktreePath),
							"Force close without committing: ticketflow close --force",
						})
				}
			}
		}
	}

	if !isWorktree {
		// Original behavior for non-worktree mode
		// Check for uncommitted changes
		if !force {
			dirty, err := app.Git.HasUncommittedChanges(ctx)
			if err != nil {
				return nil, "", fmt.Errorf("failed to check git status: %w", err)
			}
			if dirty {
				return nil, "", NewError(ErrGitDirtyWorkspace, "Uncommitted changes detected",
					"Please commit your changes before closing the ticket",
					[]string{
						"Commit your changes: git commit -am 'Your message'",
						"Force close without committing: ticketflow close --force",
					})
			}
		}

		// Get current branch
		currentBranch, err := app.Git.CurrentBranch(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get current branch: %w", err)
		}

		// Ensure we're on the ticket branch
		if currentBranch != current.ID {
			return nil, "", fmt.Errorf("not on ticket branch, expected %s but on %s", current.ID, currentBranch)
		}
	}

	return current, worktreePath, nil
}

// moveTicketToDoneWithReason moves a ticket to done and commits with optional reason
func (app *App) moveTicketToDoneWithReason(ctx context.Context, current *ticket.Ticket, reason string, isCurrentTicket bool) error {
	// Check context before starting
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Move ticket file from doing to done
	oldPath := current.Path
	donePath := app.Config.GetDonePath(app.ProjectRoot)
	newPath := filepath.Join(donePath, filepath.Base(current.Path))

	// Ensure done directory exists
	if err := os.MkdirAll(donePath, 0755); err != nil {
		return fmt.Errorf("failed to create done directory: %w", err)
	}

	// Move the file first
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to move ticket %s from %s to done: %w", current.ID, oldPath, err)
	}

	// Update ticket data with new path
	current.Path = newPath
	if err := app.Manager.Update(ctx, current); err != nil {
		// Rollback file move
		if renameErr := os.Rename(newPath, oldPath); renameErr != nil {
			return fmt.Errorf("failed to update ticket and rollback file move: %w, rename error: %v", err, renameErr)
		}
		return fmt.Errorf("failed to update ticket %s close time: %w", current.ID, err)
	}

	// Check context before git operations
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled before git operations: %w", err)
	}

	// Git add the changes (use -A to handle the rename properly)
	if err := app.Git.Add(ctx, "-A", filepath.Dir(oldPath), filepath.Dir(newPath)); err != nil {
		return fmt.Errorf("failed to stage ticket move: %w", err)
	}

	// Commit the move with reason if provided
	commitMsg := fmt.Sprintf("Close ticket: %s", current.ID)
	if reason != "" {
		commitMsg = fmt.Sprintf("Close ticket: %s (%s)", current.ID, reason)
	}
	if err := app.Git.Commit(ctx, commitMsg); err != nil {
		return fmt.Errorf("failed to commit ticket move: %w", err)
	}

	// Remove current ticket link only if this is the current ticket
	if isCurrentTicket {
		if err := app.Manager.SetCurrentTicket(ctx, nil); err != nil {
			return fmt.Errorf("failed to remove current ticket link: %w", err)
		}
	}

	return nil
}

// checkExistingWorktree checks if a worktree already exists for the ticket
func (app *App) checkExistingWorktree(ctx context.Context, t *ticket.Ticket, force bool) error {
	if !app.Config.Worktree.Enabled {
		return nil
	}

	if exists, err := app.Git.HasWorktree(ctx, t.ID); err != nil {
		return fmt.Errorf("failed to check worktree: %w", err)
	} else if exists {
		worktreePath := worktree.GetPath(ctx, app.Git, app.Config, app.ProjectRoot, t.ID)
		if !force {
			return NewError(ErrWorktreeExists, "Worktree already exists",
				fmt.Sprintf("Worktree for ticket %s already exists at: %s", t.ID, worktreePath), nil)
		}
		// Force is enabled, remove the existing worktree
		logger := log.Global().WithOperation("start_ticket").WithTicket(t.ID)
		logger.Info("removing existing worktree due to --force flag", "path", worktreePath)
		app.StatusWriter.Printf("Removing existing worktree at %s\n", worktreePath)
		if err := app.Git.RemoveWorktree(ctx, worktreePath); err != nil {
			return fmt.Errorf("failed to remove existing worktree: %w", err)
		}
	}

	return nil
}

// createAndSetupWorktree creates a worktree and runs initialization commands
func (app *App) createAndSetupWorktree(ctx context.Context, t *ticket.Ticket) (string, error) {
	logger := log.Global()

	if !app.Config.Worktree.Enabled {
		return "", nil
	}

	// Always use flat worktree structure
	baseDir := app.Config.GetWorktreePath(app.ProjectRoot)
	worktreePath := filepath.Join(baseDir, t.ID)

	err := app.Git.AddWorktree(ctx, worktreePath, t.ID)
	if err != nil {
		// Check if this is a branch divergence error
		var divergenceErr *ticketerrors.BranchDivergenceError
		if errors.As(err, &divergenceErr) {
			// Handle branch divergence
			worktreePath, err = app.handleBranchDivergence(ctx, t, worktreePath, divergenceErr)
			if err != nil {
				return "", err
			}
		} else {
			// Other error - rollback
			if _, rollbackErr := app.Git.Exec(ctx, "reset", "--hard", "HEAD^"); rollbackErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to rollback after worktree creation failure: %v\n", rollbackErr)
			}
			return "", fmt.Errorf("failed to create worktree at %s for branch %s: %w", worktreePath, t.ID, err)
		}
	}

	// Run init commands if configured
	if err := app.runWorktreeInitCommands(ctx, worktreePath); err != nil {
		// Non-fatal: just log the error
		logger.WithError(err).Warn("failed to run init commands")
		app.StatusWriter.Printf("Warning: Failed to run init commands: %v\n", err)
	}

	// Create current-ticket.md symlink in worktree
	if err := app.createWorktreeTicketSymlink(worktreePath, t); err != nil {
		return worktreePath, fmt.Errorf("failed to create current ticket link in worktree: %w", err)
	}

	return worktreePath, nil
}

// handleBranchDivergence handles the case when a branch has diverged
func (app *App) handleBranchDivergence(ctx context.Context, t *ticket.Ticket, worktreePath string,
	divergenceErr *ticketerrors.BranchDivergenceError) (string, error) {

	// Display divergence information
	app.Output.Printf("\n⚠️  Branch '%s' already exists but has diverged from '%s'\n",
		divergenceErr.Branch, divergenceErr.BaseBranch)
	app.Output.Printf("   • %d commits ahead\n", divergenceErr.Ahead)
	app.Output.Printf("   • %d commits behind\n\n", divergenceErr.Behind)

	// Prompt for user action
	options := []PromptOption{
		{Key: "u", Description: "Use existing branch as-is", IsDefault: false},
		{Key: "r", Description: "Delete and recreate branch at current HEAD", IsDefault: true},
		{Key: "c", Description: "Cancel operation", IsDefault: false},
	}

	choice, err := PromptWithStatus("How would you like to proceed?", options, app.StatusWriter)
	if err != nil {
		return "", fmt.Errorf("failed to get user choice: %w", err)
	}

	switch choice {
	case "u":
		// Use existing branch
		app.Output.Printf("Using existing branch '%s'...\n", t.ID)
		_, err = app.Git.Exec(ctx, git.SubcmdWorktree, git.WorktreeAdd, worktreePath, t.ID)
		if err != nil {
			return "", fmt.Errorf("failed to create worktree with existing branch: %w", err)
		}
		return worktreePath, nil

	case "r":
		// Delete and recreate branch
		app.Output.Printf("Recreating branch '%s' at current HEAD...\n", t.ID)

		// First, delete the branch
		_, err = app.Git.Exec(ctx, git.SubcmdBranch, git.FlagDeleteForce, t.ID)
		if err != nil {
			return "", fmt.Errorf("failed to delete branch: %w", err)
		}

		// Now create worktree with new branch
		_, err = app.Git.Exec(ctx, git.SubcmdWorktree, git.WorktreeAdd, worktreePath,
			git.FlagBranch, t.ID)
		if err != nil {
			// Try to recover by recreating the branch we just deleted
			if recoverErr := app.Git.CreateBranch(ctx, t.ID); recoverErr != nil {
				return "", fmt.Errorf("failed to create worktree with new branch and could not recover: %w (recovery error: %v)", err, recoverErr)
			}
			return "", fmt.Errorf("failed to create worktree with new branch: %w", err)
		}
		return worktreePath, nil

	case "c":
		// Cancel
		app.Output.Printf("Operation cancelled.\n")
		// Rollback the ticket start if HEAD^ exists
		if _, err := app.Git.Exec(ctx, git.SubcmdRevParse, "HEAD^"); err == nil {
			// HEAD^ exists, safe to rollback
			if _, rollbackErr := app.Git.Exec(ctx, git.SubcmdReset, git.FlagHard, "HEAD^"); rollbackErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to rollback: %v\n", rollbackErr)
			}
		}
		return "", fmt.Errorf("operation cancelled by user")

	default:
		return "", fmt.Errorf("invalid choice: %s", choice)
	}
}

// runWorktreeInitCommands runs the configured initialization commands in the worktree
func (app *App) runWorktreeInitCommands(ctx context.Context, worktreePath string) error {
	if len(app.Config.Worktree.InitCommands) == 0 {
		return nil
	}

	app.StatusWriter.Println("Running initialization commands...")
	var failedCommands []string

	// Apply timeout if not already set
	timeout := app.Config.GetInitCommandsTimeout()
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	for _, cmd := range app.Config.Worktree.InitCommands {
		app.StatusWriter.Printf("  $ %s\n", cmd)
		// Parse the command with proper shell parsing
		parts, err := shellwords.Parse(cmd)
		if err != nil {
			failedCommands = append(failedCommands, fmt.Sprintf("%s (failed to parse: %v)", cmd, err))
			continue
		}
		if len(parts) == 0 {
			continue
		}

		// Execute in worktree directory
		execCmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
		execCmd.Dir = worktreePath
		output, err := execCmd.CombinedOutput()
		if err != nil {
			// Check if error is due to timeout
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				failedCommands = append(failedCommands, fmt.Sprintf("%s (timed out after %v)", cmd, timeout))
			} else {
				failedCommands = append(failedCommands, fmt.Sprintf("%s (%v)", cmd, err))
			}
			if len(output) > 0 {
				app.StatusWriter.Printf("    Output: %s\n", strings.TrimSpace(string(output)))
			}
		}
	}

	if len(failedCommands) > 0 {
		return fmt.Errorf("some initialization commands failed: %s", strings.Join(failedCommands, ", "))
	}
	return nil
}

// createWorktreeTicketSymlink creates the current-ticket.md symlink in the worktree
func (app *App) createWorktreeTicketSymlink(worktreePath string, t *ticket.Ticket) error {
	linkPath := filepath.Join(worktreePath, "current-ticket.md")
	relPath := filepath.Join("tickets", "doing", filepath.Base(t.Path))
	return os.Symlink(relPath, linkPath)
}

// printStartSuccessMessage prints the success message after starting a ticket
func (app *App) printStartSuccessMessage(t *ticket.Ticket, worktreePath string, parentBranch string) {
	app.Output.Printf("\n✅ Started work on ticket: %s\n", t.ID)
	app.Output.Printf("   Description: %s\n", t.Description)

	if app.Config.Worktree.Enabled {
		app.printWorktreeStartMessage(t, worktreePath, parentBranch)
	} else {
		app.printBranchStartMessage(t)
	}
}

// printWorktreeStartMessage prints the success message for worktree mode
func (app *App) printWorktreeStartMessage(t *ticket.Ticket, worktreePath string, parentBranch string) {
	app.Output.Printf("\n📁 Worktree created: %s\n", worktreePath)
	if parentBranch != "" {
		app.Output.Printf("   Parent ticket: %s\n", parentBranch)
		app.Output.Printf("   Branch from: %s\n", parentBranch)
	}
	app.Output.Printf("   Status: todo → doing\n")
	app.Output.Printf("   Committed: \"Start ticket: %s\"\n", t.ID)
	app.Output.Printf("\n📋 Next steps:\n")
	app.Output.Printf("1. Navigate to worktree:\n")
	app.Output.Printf("   cd %s\n", worktreePath)
	app.Output.Printf("   \n")
	app.Output.Printf("2. Make your changes and commit regularly\n")
	app.Output.Printf("   \n")
	app.Output.Printf("3. Push branch to create PR:\n")
	app.Output.Printf("   git push -u origin %s\n", t.ID)
	app.Output.Printf("   \n")
	app.Output.Printf("4. When done, close the ticket:\n")
	app.Output.Printf("   ticketflow close\n")
}

// printBranchStartMessage prints the success message for non-worktree mode
func (app *App) printBranchStartMessage(t *ticket.Ticket) {
	app.Output.Printf("\n🌿 Switched to branch: %s\n", t.ID)
	app.Output.Printf("   Status: todo → doing\n")
	app.Output.Printf("   Committed: \"Start ticket: %s\"\n", t.ID)
	app.Output.Printf("\n📋 Next steps:\n")
	app.Output.Printf("1. Make your changes and commit regularly\n")
	app.Output.Printf("   \n")
	app.Output.Printf("2. Push branch to create PR:\n")
	app.Output.Printf("   git push -u origin %s\n", t.ID)
	app.Output.Printf("   \n")
	app.Output.Printf("3. When done, close the ticket:\n")
	app.Output.Printf("   ticketflow close\n")
}

// calculateWorkDuration calculates the work duration for a closed ticket
func (app *App) calculateWorkDuration(t *ticket.Ticket) string {
	if t.StartedAt.Time != nil && t.ClosedAt.Time != nil {
		dur := t.ClosedAt.Time.Sub(*t.StartedAt.Time)
		return formatDuration(dur)
	}
	return ""
}

// extractParentTicketID extracts the parent ticket ID from the related field
func (app *App) extractParentTicketID(t *ticket.Ticket) string {
	for _, rel := range t.Related {
		if strings.HasPrefix(rel, "parent:") {
			return strings.TrimPrefix(rel, "parent:")
		}
	}
	return ""
}

// printCloseSuccessMessage prints the success message after closing a ticket
func (app *App) printCloseSuccessMessage(t *ticket.Ticket, duration, parentTicketID, worktreePath string) {
	app.Output.Printf("\n✅ Ticket closed: %s\n", t.ID)
	app.Output.Printf("   Description: %s\n", t.Description)
	app.Output.Printf("   Status: doing → done\n")
	if duration != "" {
		app.Output.Printf("   Duration: %s\n", duration)
	}
	app.Output.Printf("   Committed: \"Close ticket: %s\"\n", t.ID)

	if parentTicketID != "" {
		app.Output.Printf("   Parent ticket: %s\n", parentTicketID)
	}

	app.Output.Printf("\n📋 Next steps:\n")
	app.Output.Printf("1. Push your branch to create/update PR:\n")
	app.Output.Printf("   git push origin %s\n", t.ID)
	app.Output.Printf("   \n")
	app.Output.Printf("2. Create Pull Request on your Git service\n")
	app.Output.Printf("   - Title: %s\n", t.Description)
	app.Output.Printf("   - Target: %s\n", app.Config.Git.DefaultBranch)
	if parentTicketID != "" {
		app.Output.Printf("   - Mention parent ticket: %s\n", parentTicketID)
	}
	app.Output.Printf("   \n")
	app.Output.Printf("3. After PR is merged, clean up:\n")
	app.Output.Printf("   ticketflow cleanup %s\n", t.ID)

	if worktreePath != "" {
		app.Output.Printf("\n🌳 Note: Worktree remains at %s\n", worktreePath)
		app.Output.Printf("   You can continue working there until cleanup\n")
	}
}

// countTicketsByStatus counts tickets by their status
func (app *App) countTicketsByStatus(tickets []ticket.Ticket) (todoCount, doingCount, doneCount int) {
	for _, t := range tickets {
		switch t.Status() {
		case ticket.StatusTodo:
			todoCount++
		case ticket.StatusDoing:
			doingCount++
		case ticket.StatusDone:
			doneCount++
		}
	}
	return todoCount, doingCount, doneCount
}

// formatStatusJSON formats the status output as JSON
func (app *App) formatStatusJSON(branch string, current *ticket.Ticket, allTickets []ticket.Ticket, todoCount, doingCount, doneCount int) error {
	output := map[string]interface{}{
		"current_branch": branch,
		"summary": map[string]int{
			"total": len(allTickets),
			"todo":  todoCount,
			"doing": doingCount,
			"done":  doneCount,
		},
	}

	if current != nil {
		output["current_ticket"] = ticketToJSON(current, "")
	} else {
		output["current_ticket"] = nil
	}

	return app.Output.PrintJSON(output)
}

// printStatusText prints the status in text format
func (app *App) printStatusText(ctx context.Context, branch string, current *ticket.Ticket, allTickets []ticket.Ticket, todoCount, doingCount, doneCount int) {
	app.Output.Printf("\n🌿 Current branch: %s\n", branch)

	if current != nil {
		app.Output.Printf("\n🎯 Active ticket: %s\n", current.ID)
		app.Output.Printf("   Description: %s\n", current.Description)
		app.Output.Printf("   Status: %s\n", current.Status())
		if current.StartedAt.Time != nil {
			duration := time.Since(*current.StartedAt.Time)
			app.Output.Printf("   Duration: %s\n", formatDuration(duration))
		}

		// Check if in worktree
		if app.Config.Worktree.Enabled {
			wt, _ := app.Git.FindWorktreeByBranch(ctx, current.ID)
			if wt != nil {
				app.Output.Printf("   Worktree: %s\n", wt.Path)
			}
		}
	} else {
		app.Output.Println("\n⚠️  No active ticket")
		app.Output.Println("   Start a ticket with: ticketflow start <ticket-id>")
	}

	app.Output.Printf("\n📊 Ticket summary:\n")
	app.Output.Printf("   📘 Todo:  %d\n", todoCount)
	app.Output.Printf("   🔨 Doing: %d\n", doingCount)
	app.Output.Printf("   ✅ Done:  %d\n", doneCount)
	app.Output.Printf("   ─────────\n")
	app.Output.Printf("   🔢 Total: %d\n", len(allTickets))
}
