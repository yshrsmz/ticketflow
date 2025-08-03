package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-shellwords"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/log"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// App represents the CLI application
type App struct {
	Config      *config.Config
	Git         git.GitClient
	Manager     ticket.TicketManager
	ProjectRoot string
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

// NewAppWithOptions creates a new CLI application with custom options
func NewAppWithOptions(ctx context.Context, opts ...AppOption) (*App, error) {
	// Find project root (with .git directory)
	projectRoot, err := git.FindProjectRoot(ctx, ".")
	if err != nil {
		return nil, NewError(ErrNotGitRepo, "Not in a git repository", "",
			[]string{
				"Navigate to your project root directory",
				"Initialize a new git repository with 'git init'",
			})
	}

	// Load config
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, NewError(ErrConfigNotFound, "Ticket system not initialized", "",
			[]string{
				"Run 'ticketflow init' to initialize",
				"Navigate to the project root directory",
			})
	}

	app := &App{
		Config:      cfg,
		ProjectRoot: projectRoot,
	}

	// Apply options
	for _, opt := range opts {
		opt(app)
	}

	// Set defaults if not provided
	if app.Git == nil {
		app.Git = git.NewWithTimeout(projectRoot, app.Config.GetGitTimeout())
	}
	if app.Manager == nil {
		app.Manager = ticket.NewManager(cfg, projectRoot)
	}

	return app, nil
}

// NewApp creates a new CLI application
func NewApp(ctx context.Context) (*App, error) {
	// Find project root (with .git directory)
	projectRoot, err := git.FindProjectRoot(ctx, ".")
	if err != nil {
		return nil, NewError(ErrNotGitRepo, "Not in a git repository", "",
			[]string{
				"Navigate to your project root directory",
				"Initialize a new git repository with 'git init'",
			})
	}

	// Load config
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, NewError(ErrConfigNotFound, "Ticket system not initialized", "",
			[]string{
				"Run 'ticketflow init' to initialize",
				"Navigate to the project root directory",
			})
	}

	gitClient := git.NewWithTimeout(projectRoot, cfg.GetGitTimeout())
	manager := ticket.NewManager(cfg, projectRoot)

	return &App{
		Config:      cfg,
		Git:         gitClient,
		Manager:     manager,
		ProjectRoot: projectRoot,
	}, nil
}

// InitCommand initializes the ticket system (doesn't require existing config)
func InitCommand(ctx context.Context) error {
	logger := log.Global().WithOperation("init")

	projectRoot, err := git.FindProjectRoot(ctx, ".")
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
		fmt.Println("Ticket system already initialized")
		return nil
	}

	// Save config
	if err := cfg.Save(configPath); err != nil {
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
	fmt.Println("Initialized ticket system successfully")
	fmt.Printf("Configuration saved to: %s\n", configPath)
	fmt.Printf("Tickets directory: %s\n", ticketsDir)

	return nil
}

// NewTicket creates a new ticket
func (app *App) NewTicket(ctx context.Context, slug string, format OutputFormat) error {
	logger := log.Global().WithOperation("new_ticket")

	// Validate slug
	if !ticket.IsValidSlug(slug) {
		logger.Error("invalid slug format", slog.String("slug", slug))
		return NewError(ErrTicketInvalid, "Invalid slug format",
			fmt.Sprintf("Slug '%s' contains invalid characters", slug),
			[]string{
				"Use only lowercase letters (a-z)",
				"Use only numbers (0-9)",
				"Use only hyphens (-) for separation",
			})
	}
	logger.Debug("creating new ticket", "slug", slug)

	// Auto-detect parent ticket from current branch
	var parentTicketID string
	currentBranch, err := app.Git.CurrentBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// If we're not on the default branch, check if it's a ticket branch
	if currentBranch != app.Config.Git.DefaultBranch {
		// Verify this is a valid ticket
		if _, err := app.Manager.Get(ctx, currentBranch); err == nil {
			parentTicketID = currentBranch
			fmt.Printf("Creating ticket in branch: %s\n", currentBranch)
		}
	}

	// Create ticket
	t, err := app.Manager.Create(ctx, slug)
	if err != nil {
		logger.WithError(err).Error("failed to create ticket", slog.String("slug", slug))
		return ConvertError(err)
	}
	logger.Info("created ticket", "ticket_id", t.ID, "path", t.Path)

	// If this is a sub-ticket, update its metadata
	if parentTicketID != "" {
		logger.Debug("creating sub-ticket", "parent", parentTicketID)
		// Add parent relationship
		t.Related = append(t.Related, fmt.Sprintf("parent:%s", parentTicketID))
		if err := app.Manager.Update(ctx, t); err != nil {
			logger.WithError(err).Error("failed to update ticket metadata", slog.String("ticket_id", t.ID), slog.String("parent", parentTicketID))
			return fmt.Errorf("failed to update ticket metadata: %w", err)
		}
		logger.Info("created sub-ticket", "ticket_id", t.ID, "parent", parentTicketID)
	}

	if format == FormatJSON {
		output := map[string]interface{}{
			"ticket": map[string]interface{}{
				"id":   t.ID,
				"path": t.Path,
			},
		}
		if parentTicketID != "" {
			output["parent_ticket"] = parentTicketID
		}
		return outputJSON(output)
	}

	fmt.Printf("\n🎫 Created new ticket: %s\n", t.ID)
	fmt.Printf("   File: %s\n", t.Path)
	if parentTicketID != "" {
		fmt.Printf("   Parent ticket: %s\n", parentTicketID)
		fmt.Printf("   Type: Sub-ticket\n")
	}
	fmt.Printf("\n📋 Next steps:\n")
	fmt.Printf("1. Edit the ticket file to add details:\n")
	fmt.Printf("   $EDITOR %s\n", t.Path)
	fmt.Printf("   \n")
	fmt.Printf("2. Commit the ticket file:\n")
	fmt.Printf("   git add %s\n", t.Path)
	fmt.Printf("   git commit -m \"Add ticket: %s\"\n", slug)
	fmt.Printf("   \n")
	fmt.Printf("3. Start working on it:\n")
	fmt.Printf("   ticketflow start %s\n", t.ID)

	return nil
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

	// Convert to pointer slice for output functions
	ticketPtrs := make([]*ticket.Ticket, len(tickets))
	for i := range tickets {
		ticketPtrs[i] = &tickets[i]
	}

	if format == FormatJSON {
		return app.outputTicketListJSON(ctx, ticketPtrs)
	}

	return app.outputTicketListText(ticketPtrs)
}

// StartTicket starts working on a ticket
func (app *App) StartTicket(ctx context.Context, ticketID string) error {
	logger := log.Global().WithOperation("start_ticket").WithTicket(ticketID)
	logger.Info("starting ticket")

	// Get and validate the ticket
	t, err := app.validateTicketForStart(ctx, ticketID)
	if err != nil {
		logger.WithError(err).Error("failed to validate ticket")
		return err
	}

	// Check workspace state
	if err := app.checkWorkspaceForStart(ctx); err != nil {
		return err
	}

	// Get current branch and detect parent
	currentBranch, parentBranch, err := app.detectParentBranch(ctx)
	if err != nil {
		return err
	}

	// Setup branch for the ticket
	if err := app.setupTicketBranch(ctx, t, currentBranch); err != nil {
		return err
	}

	// Check if worktree already exists (for worktree mode)
	if err := app.checkExistingWorktree(ctx, t); err != nil {
		return err
	}

	// Update parent relationship if needed
	if err := app.updateParentRelationship(ctx, t, parentBranch, currentBranch); err != nil {
		return err
	}

	// Move ticket to doing status
	if err := app.moveTicketToDoing(ctx, t, currentBranch); err != nil {
		return err
	}

	// Now create worktree AFTER committing (for worktree mode)
	worktreePath, err := app.createAndSetupWorktree(ctx, t)
	if err != nil {
		return err
	}

	// Output success message
	app.printStartSuccessMessage(t, worktreePath, parentBranch)

	return nil
}

// CloseTicket closes the current ticket
func (app *App) CloseTicket(ctx context.Context, force bool) error {
	logger := log.Global().WithOperation("close_ticket")

	// Validate current ticket for close
	current, worktreePath, err := app.validateTicketForClose(ctx, force)
	if err != nil {
		logger.WithError(err).Error("failed to validate ticket for close")
		return err
	}

	logger = logger.WithTicket(current.ID)
	logger.Info("closing ticket")

	// Update ticket status
	if err := current.Close(); err != nil {
		return err
	}

	// Move ticket to done status
	if err := app.moveTicketToDone(ctx, current); err != nil {
		return err
	}

	// Calculate work duration
	duration := app.calculateWorkDuration(current)

	// Extract parent ticket ID
	parentTicketID := app.extractParentTicketID(current)

	// Print success message with next steps
	app.printCloseSuccessMessage(current, duration, parentTicketID, worktreePath)

	logger.Info("ticket closed successfully", "duration", duration)
	return nil
}

// RestoreCurrentTicket restores the current ticket symlink
func (app *App) RestoreCurrentTicket(ctx context.Context) error {
	// Get current branch
	branch, err := app.Git.CurrentBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Try to get ticket by branch name
	t, err := app.Manager.Get(ctx, branch)
	if err != nil {
		return ConvertError(fmt.Errorf("no ticket found for branch %s", branch))
	}

	// Set current ticket
	if err := app.Manager.SetCurrentTicket(ctx, t); err != nil {
		return fmt.Errorf("failed to set current ticket: %w", err)
	}

	fmt.Printf("Restored current ticket link: %s\n", t.ID)
	return nil
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
		fmt.Println("No tickets found")
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
	fmt.Printf("%-*s  %-6s  %-3s  %s\n", maxIDLen, "ID", "STATUS", "PRI", "DESCRIPTION")
	fmt.Println(strings.Repeat("-", maxIDLen+50))

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

		fmt.Printf("%-*s  %-6s  %-3s  %s\n", maxIDLen, t.ID, status, priority, desc)
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

	return outputJSON(output)
}

// ListWorktrees lists all worktrees
func (app *App) ListWorktrees(ctx context.Context, format OutputFormat) error {
	worktrees, err := app.Git.ListWorktrees(ctx)
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	if format == FormatJSON {
		output := map[string]interface{}{
			"worktrees": worktrees,
		}
		return outputJSON(output)
	}

	// Text format
	if len(worktrees) == 0 {
		fmt.Println("No worktrees found")
		return nil
	}

	fmt.Printf("%-50s %-30s %s\n", "PATH", "BRANCH", "HEAD")
	fmt.Println(strings.Repeat("-", 100))

	for _, wt := range worktrees {
		head := wt.HEAD
		if len(head) > 40 {
			head = head[:7] // Short commit hash
		}
		fmt.Printf("%-50s %-30s %s\n", wt.Path, wt.Branch, head)
	}

	return nil
}

// CleanWorktrees removes orphaned worktrees
func (app *App) CleanWorktrees(ctx context.Context) error {
	logger := log.Global()

	// First prune to clean up git's internal state
	if err := app.Git.PruneWorktrees(ctx); err != nil {
		return fmt.Errorf("failed to prune worktrees: %w", err)
	}

	// Get all worktrees
	worktrees, err := app.Git.ListWorktrees(ctx)
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Get all active tickets
	activeTickets, err := app.Manager.List(ctx, ticket.StatusFilterDoing)
	if err != nil {
		return fmt.Errorf("failed to list active tickets: %w", err)
	}

	// Create a map of active ticket IDs
	activeMap := make(map[string]bool)
	for _, t := range activeTickets {
		activeMap[t.ID] = true
	}

	// Find and remove orphaned worktrees
	cleaned := 0
	for _, wt := range worktrees {
		// Skip main worktree
		if wt.Branch == "" || wt.Branch == app.Config.Git.DefaultBranch {
			continue
		}

		// Check if this worktree corresponds to an active ticket
		if !activeMap[wt.Branch] {
			fmt.Printf("Removing orphaned worktree: %s (branch: %s)\n", wt.Path, wt.Branch)
			if err := app.Git.RemoveWorktree(ctx, wt.Path); err != nil {
				logger.WithError(err).Warn("failed to remove worktree", "path", wt.Path)
				fmt.Printf("Warning: Failed to remove worktree: %v\n", err)
			} else {
				cleaned++
			}
		}
	}

	if cleaned == 0 {
		fmt.Println("No orphaned worktrees found")
	} else {
		fmt.Printf("Cleaned %d orphaned worktree(s)\n", cleaned)
	}

	return nil
}

// CleanupTicket cleans up a ticket after PR merge
func (app *App) CleanupTicket(ctx context.Context, ticketID string, force bool) error {
	// Get the ticket to verify it exists and is done
	t, err := app.Manager.Get(ctx, ticketID)
	if err != nil {
		return ConvertError(err)
	}

	// Check if ticket is done
	if t.Status() != ticket.StatusDone {
		return NewError(ErrTicketNotDone, "Ticket is not done",
			fmt.Sprintf("Ticket %s is in '%s' status, not 'done'", t.ID, t.Status()),
			[]string{
				"Close the ticket first: ticketflow close",
				"Or manually move the ticket to done directory",
			})
	}

	// Get current branch to restore later
	currentBranch, err := app.Git.CurrentBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Make sure we're on the default branch
	if currentBranch != app.Config.Git.DefaultBranch {
		if err := app.Git.Checkout(ctx, app.Config.Git.DefaultBranch); err != nil {
			return fmt.Errorf("failed to checkout default branch: %w", err)
		}
	}

	// Check if worktree exists
	wt, err := app.Git.FindWorktreeByBranch(ctx, t.ID)
	if err != nil {
		return fmt.Errorf("failed to find worktree: %w", err)
	}

	// Show what will be done
	fmt.Printf("\n🗑️  Cleanup for ticket: %s\n", t.ID)
	fmt.Printf("   Description: %s\n", t.Description)
	fmt.Printf("\nThis will:\n")
	if wt != nil {
		fmt.Printf("  • Remove worktree: %s\n", wt.Path)
	}
	fmt.Printf("  • Delete local branch: %s\n", t.ID)

	// Confirmation prompt if not forced
	if !force {
		fmt.Printf("\nAre you sure? (y/N): ")

		var response string
		_, _ = fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("\n❌ Cleanup cancelled")
			return nil
		}
	}

	fmt.Printf("\n🔧 Performing cleanup...\n")

	// Remove worktree if it exists
	if wt != nil {
		fmt.Printf("🌳 Removing worktree: %s\n", wt.Path)
		if err := app.Git.RemoveWorktree(ctx, wt.Path); err != nil {
			return fmt.Errorf("failed to remove worktree at %s for ticket %s: %w", wt.Path, ticketID, err)
		}
	}

	// Delete local branch
	fmt.Printf("🌿 Deleting local branch: %s\n", t.ID)
	if _, err := app.Git.Exec(ctx, "branch", "-D", t.ID); err != nil {
		// Branch might not exist locally, which is fine
		fmt.Printf("⚠️  Note: Local branch %s not found or already deleted\n", t.ID)
	}

	fmt.Printf("\n✅ Cleanup completed successfully!\n")
	fmt.Printf("\n📋 What's next:\n")
	fmt.Printf("• Start a new ticket: ticketflow new <slug>\n")
	fmt.Printf("• View open tickets: ticketflow list --status todo\n")
	fmt.Printf("• Check active work: ticketflow list --status doing\n")
	return nil
}

// validateTicketForStart validates that a ticket can be started
func (app *App) validateTicketForStart(ctx context.Context, ticketID string) (*ticket.Ticket, error) {
	// Get the ticket
	t, err := app.Manager.Get(ctx, ticketID)
	if err != nil {
		return nil, ConvertError(err)
	}

	// Check if already started
	if t.Status() == ticket.StatusDoing {
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

	// Set as current ticket
	if err := app.Manager.SetCurrentTicket(ctx, t); err != nil {
		return fmt.Errorf("failed to set current ticket: %w", err)
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
		return nil, "", ConvertError(err)
	}
	if current == nil {
		return nil, "", NewError(ErrTicketNotStarted, "No active ticket",
			"There is no ticket currently being worked on",
			[]string{
				"Start a ticket first: ticketflow start <ticket-id>",
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

// moveTicketToDone moves the ticket to done status and commits the change
func (app *App) moveTicketToDone(ctx context.Context, current *ticket.Ticket) error {
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

	// Git add the changes (use -A to handle the rename properly)
	if err := app.Git.Add(ctx, "-A", filepath.Dir(oldPath), filepath.Dir(newPath)); err != nil {
		return fmt.Errorf("failed to stage ticket move: %w", err)
	}

	// Commit the move
	if err := app.Git.Commit(ctx, fmt.Sprintf("Close ticket: %s", current.ID)); err != nil {
		return fmt.Errorf("failed to commit ticket move: %w", err)
	}

	// Remove current ticket link
	if err := app.Manager.SetCurrentTicket(ctx, nil); err != nil {
		return fmt.Errorf("failed to remove current ticket link: %w", err)
	}

	return nil
}

// getWorktreePath attempts to get the actual worktree path or falls back to calculated path
func (app *App) getWorktreePath(ctx context.Context, ticketID string) string {
	logger := log.Global().WithTicket(ticketID)
	
	// Try to get the actual worktree path
	if wt, err := app.Git.FindWorktreeByBranch(ctx, ticketID); err == nil && wt != nil {
		logger.Debug("found worktree path", "path", wt.Path)
		return wt.Path
	}
	
	// Fall back to calculated path
	baseDir := app.Config.GetWorktreePath(app.ProjectRoot)
	calculatedPath := filepath.Join(baseDir, ticketID)
	logger.Debug("using calculated worktree path", "path", calculatedPath)
	return calculatedPath
}

// checkExistingWorktree checks if a worktree already exists for the ticket
func (app *App) checkExistingWorktree(ctx context.Context, t *ticket.Ticket) error {
	if !app.Config.Worktree.Enabled {
		return nil
	}

	if exists, err := app.Git.HasWorktree(ctx, t.ID); err != nil {
		return fmt.Errorf("failed to check worktree: %w", err)
	} else if exists {
		worktreePath := app.getWorktreePath(ctx, t.ID)
		return NewError(ErrWorktreeExists, "Worktree already exists",
			fmt.Sprintf("Worktree for ticket %s already exists at: %s", t.ID, worktreePath), nil)
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

	if err := app.Git.AddWorktree(ctx, worktreePath, t.ID); err != nil {
		// Rollback: reset to previous commit
		if _, rollbackErr := app.Git.Exec(ctx, "reset", "--hard", "HEAD^"); rollbackErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to rollback after worktree creation failure: %v\n", rollbackErr)
		}
		return "", fmt.Errorf("failed to create worktree at %s for branch %s: %w", worktreePath, t.ID, err)
	}

	// Run init commands if configured
	if err := app.runWorktreeInitCommands(ctx, worktreePath); err != nil {
		// Non-fatal: just log the error
		logger.WithError(err).Warn("failed to run init commands")
		fmt.Printf("Warning: Failed to run init commands: %v\n", err)
	}

	// Create current-ticket.md symlink in worktree
	if err := app.createWorktreeTicketSymlink(worktreePath, t); err != nil {
		return worktreePath, fmt.Errorf("failed to create current ticket link in worktree: %w", err)
	}

	return worktreePath, nil
}

// runWorktreeInitCommands runs the configured initialization commands in the worktree
func (app *App) runWorktreeInitCommands(ctx context.Context, worktreePath string) error {
	if len(app.Config.Worktree.InitCommands) == 0 {
		return nil
	}

	fmt.Println("Running initialization commands...")
	var failedCommands []string

	// Apply timeout if not already set
	timeout := app.Config.GetInitCommandsTimeout()
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	for _, cmd := range app.Config.Worktree.InitCommands {
		fmt.Printf("  $ %s\n", cmd)
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
				fmt.Printf("    Output: %s\n", strings.TrimSpace(string(output)))
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
	fmt.Printf("\n✅ Started work on ticket: %s\n", t.ID)
	fmt.Printf("   Description: %s\n", t.Description)

	if app.Config.Worktree.Enabled {
		app.printWorktreeStartMessage(t, worktreePath, parentBranch)
	} else {
		app.printBranchStartMessage(t)
	}
}

// printWorktreeStartMessage prints the success message for worktree mode
func (app *App) printWorktreeStartMessage(t *ticket.Ticket, worktreePath string, parentBranch string) {
	fmt.Printf("\n📁 Worktree created: %s\n", worktreePath)
	if parentBranch != "" {
		fmt.Printf("   Parent ticket: %s\n", parentBranch)
		fmt.Printf("   Branch from: %s\n", parentBranch)
	}
	fmt.Printf("   Status: todo → doing\n")
	fmt.Printf("   Committed: \"Start ticket: %s\"\n", t.ID)
	fmt.Printf("\n📋 Next steps:\n")
	fmt.Printf("1. Navigate to worktree:\n")
	fmt.Printf("   cd %s\n", worktreePath)
	fmt.Printf("   \n")
	fmt.Printf("2. Make your changes and commit regularly\n")
	fmt.Printf("   \n")
	fmt.Printf("3. Push branch to create PR:\n")
	fmt.Printf("   git push -u origin %s\n", t.ID)
	fmt.Printf("   \n")
	fmt.Printf("4. When done, close the ticket:\n")
	fmt.Printf("   ticketflow close\n")
}

// printBranchStartMessage prints the success message for non-worktree mode
func (app *App) printBranchStartMessage(t *ticket.Ticket) {
	fmt.Printf("\n🌿 Switched to branch: %s\n", t.ID)
	fmt.Printf("   Status: todo → doing\n")
	fmt.Printf("   Committed: \"Start ticket: %s\"\n", t.ID)
	fmt.Printf("\n📋 Next steps:\n")
	fmt.Printf("1. Make your changes and commit regularly\n")
	fmt.Printf("   \n")
	fmt.Printf("2. Push branch to create PR:\n")
	fmt.Printf("   git push -u origin %s\n", t.ID)
	fmt.Printf("   \n")
	fmt.Printf("3. When done, close the ticket:\n")
	fmt.Printf("   ticketflow close\n")
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
	fmt.Printf("\n✅ Ticket closed: %s\n", t.ID)
	fmt.Printf("   Description: %s\n", t.Description)
	fmt.Printf("   Status: doing → done\n")
	if duration != "" {
		fmt.Printf("   Duration: %s\n", duration)
	}
	fmt.Printf("   Committed: \"Close ticket: %s\"\n", t.ID)

	if parentTicketID != "" {
		fmt.Printf("   Parent ticket: %s\n", parentTicketID)
	}

	fmt.Printf("\n📋 Next steps:\n")
	fmt.Printf("1. Push your branch to create/update PR:\n")
	fmt.Printf("   git push origin %s\n", t.ID)
	fmt.Printf("   \n")
	fmt.Printf("2. Create Pull Request on your Git service\n")
	fmt.Printf("   - Title: %s\n", t.Description)
	fmt.Printf("   - Target: %s\n", app.Config.Git.DefaultBranch)
	if parentTicketID != "" {
		fmt.Printf("   - Mention parent ticket: %s\n", parentTicketID)
	}
	fmt.Printf("   \n")
	fmt.Printf("3. After PR is merged, clean up:\n")
	fmt.Printf("   ticketflow cleanup %s\n", t.ID)

	if worktreePath != "" {
		fmt.Printf("\n🌳 Note: Worktree remains at %s\n", worktreePath)
		fmt.Printf("   You can continue working there until cleanup\n")
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

	return outputJSON(output)
}

// printStatusText prints the status in text format
func (app *App) printStatusText(ctx context.Context, branch string, current *ticket.Ticket, allTickets []ticket.Ticket, todoCount, doingCount, doneCount int) {
	fmt.Printf("\n🌿 Current branch: %s\n", branch)

	if current != nil {
		fmt.Printf("\n🎯 Active ticket: %s\n", current.ID)
		fmt.Printf("   Description: %s\n", current.Description)
		fmt.Printf("   Status: %s\n", current.Status())
		if current.StartedAt.Time != nil {
			duration := time.Since(*current.StartedAt.Time)
			fmt.Printf("   Duration: %s\n", formatDuration(duration))
		}

		// Check if in worktree
		if app.Config.Worktree.Enabled {
			wt, _ := app.Git.FindWorktreeByBranch(ctx, current.ID)
			if wt != nil {
				fmt.Printf("   Worktree: %s\n", wt.Path)
			}
		}
	} else {
		fmt.Println("\n⚠️  No active ticket")
		fmt.Println("   Start a ticket with: ticketflow start <ticket-id>")
	}

	fmt.Printf("\n📊 Ticket summary:\n")
	fmt.Printf("   📘 Todo:  %d\n", todoCount)
	fmt.Printf("   🔨 Doing: %d\n", doingCount)
	fmt.Printf("   ✅ Done:  %d\n", doneCount)
	fmt.Printf("   ─────────\n")
	fmt.Printf("   🔢 Total: %d\n", len(allTickets))
}
