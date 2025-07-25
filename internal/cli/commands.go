package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// App represents the CLI application
type App struct {
	Config      *config.Config
	Git         *git.Git
	Manager     *ticket.Manager
	ProjectRoot string
}

// NewApp creates a new CLI application
func NewApp() (*App, error) {
	// Find project root (with .git directory)
	projectRoot, err := git.FindProjectRoot(".")
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

	gitClient := git.New(projectRoot)
	manager := ticket.NewManager(cfg, projectRoot)

	return &App{
		Config:      cfg,
		Git:         gitClient,
		Manager:     manager,
		ProjectRoot: projectRoot,
	}, nil
}

// InitCommand initializes the ticket system (doesn't require existing config)
func InitCommand() error {
	projectRoot, err := git.FindProjectRoot(".")
	if err != nil {
		return NewError(ErrNotGitRepo, "Not in a git repository", "", nil)
	}

	// Create default config
	cfg := config.Default()
	configPath := filepath.Join(projectRoot, ".ticketflow.yaml")

	// Check if already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("Ticket system already initialized")
		return nil
	}

	// Save config
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Create directory structure
	ticketsDir := filepath.Join(projectRoot, cfg.Tickets.Dir)
	todoDir := filepath.Join(ticketsDir, cfg.Tickets.TodoDir)
	doingDir := filepath.Join(ticketsDir, cfg.Tickets.DoingDir)
	doneDir := filepath.Join(ticketsDir, cfg.Tickets.DoneDir)
	
	// Create all directories
	for _, dir := range []string{ticketsDir, todoDir, doingDir, doneDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Update .gitignore
	gitignorePath := filepath.Join(projectRoot, ".gitignore")
	if err := updateGitignore(gitignorePath); err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}

	fmt.Println("Initialized ticket system successfully")
	fmt.Printf("Configuration saved to: %s\n", configPath)
	fmt.Printf("Tickets directory: %s\n", ticketsDir)

	return nil
}

// NewTicket creates a new ticket
func (app *App) NewTicket(slug string, format OutputFormat) error {
	// Validate slug
	if !ticket.IsValidSlug(slug) {
		return NewError(ErrTicketInvalid, "Invalid slug format",
			fmt.Sprintf("Slug '%s' contains invalid characters", slug),
			[]string{
				"Use only lowercase letters (a-z)",
				"Use only numbers (0-9)",
				"Use only hyphens (-) for separation",
			})
	}

	// Create ticket
	t, err := app.Manager.Create(slug)
	if err != nil {
		return err
	}

	if format == FormatJSON {
		return outputJSON(map[string]interface{}{
			"ticket": map[string]interface{}{
				"id":   t.ID,
				"path": t.Path,
			},
		})
	}

	fmt.Printf("Created ticket file: %s\n", t.Path)
	fmt.Println("Please edit the file to add title, description and details.")

	return nil
}

// ListTickets lists tickets
func (app *App) ListTickets(status ticket.Status, count int, format OutputFormat) error {
	tickets, err := app.Manager.List(string(status))
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
		return app.outputTicketListJSON(ticketPtrs)
	}

	return app.outputTicketListText(ticketPtrs)
}

// StartTicket starts working on a ticket
func (app *App) StartTicket(ticketID string, noPush bool) error {
	// Get the ticket
	t, err := app.Manager.Get(ticketID)
	if err != nil {
		return err
	}

	// Check if already started
	if t.Status() == ticket.StatusDoing {
		return NewError(ErrTicketAlreadyStarted, "Ticket already started",
			fmt.Sprintf("Ticket %s is already in progress", t.ID), nil)
	}

	// Check for uncommitted changes (only if not using worktrees)
	if !app.Config.Worktree.Enabled {
		dirty, err := app.Git.HasUncommittedChanges()
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

	// Get current branch
	currentBranch, err := app.Git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	var worktreePath string

	if app.Config.Worktree.Enabled {
		// Check if worktree already exists
		if exists, err := app.Git.HasWorktree(t.ID); err != nil {
			return fmt.Errorf("failed to check worktree: %w", err)
		} else if exists {
			return NewError(ErrWorktreeExists, "Worktree already exists",
				fmt.Sprintf("Worktree for ticket %s already exists", t.ID), nil)
		}

		// Create worktree
		baseDir := app.Config.GetWorktreePath(app.ProjectRoot)
		worktreePath = filepath.Join(baseDir, t.ID)
		
		if err := app.Git.AddWorktree(worktreePath, t.ID); err != nil {
			return fmt.Errorf("failed to create worktree: %w", err)
		}

		// Run init commands if configured
		if app.Config.Worktree.AutoOperations.CreateOnStart && len(app.Config.Worktree.InitCommands) > 0 {
			fmt.Println("Running initialization commands...")
			for _, cmd := range app.Config.Worktree.InitCommands {
				fmt.Printf("  $ %s\n", cmd)
				// Parse the command
				parts := strings.Fields(cmd)
				if len(parts) == 0 {
					continue
				}
				
				// Execute in worktree directory
				execCmd := exec.Command(parts[0], parts[1:]...)
				execCmd.Dir = worktreePath
				output, err := execCmd.CombinedOutput()
				if err != nil {
					fmt.Printf("Warning: Command failed: %v\n%s\n", err, output)
				}
			}
		}
	} else {
		// Original behavior: create and checkout branch
		if err := app.Git.CreateBranch(t.ID); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
	}

	// Update ticket status
	if err := t.Start(); err != nil {
		// Rollback
		if app.Config.Worktree.Enabled {
			app.Git.RemoveWorktree(worktreePath)
		} else {
			app.Git.Checkout(currentBranch)
		}
		return err
	}

	// Move ticket file from todo to doing
	oldPath := t.Path
	doingPath := app.Config.GetDoingPath(app.ProjectRoot)
	newPath := filepath.Join(doingPath, filepath.Base(t.Path))
	
	// Move the file first
	if err := os.Rename(oldPath, newPath); err != nil {
		// Rollback
		if app.Config.Worktree.Enabled {
			app.Git.RemoveWorktree(worktreePath)
		} else {
			app.Git.Checkout(currentBranch)
		}
		return fmt.Errorf("failed to move ticket to doing: %w", err)
	}
	
	// Update ticket data with new path
	t.Path = newPath
	if err := app.Manager.Update(t); err != nil {
		// Rollback file move
		os.Rename(newPath, oldPath)
		if app.Config.Worktree.Enabled {
			app.Git.RemoveWorktree(worktreePath)
		} else {
			app.Git.Checkout(currentBranch)
		}
		return fmt.Errorf("failed to update ticket: %w", err)
	}
	
	// Git add both old and new paths
	if err := app.Git.Add(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to stage ticket move: %w", err)
	}
	
	// Commit the move
	if err := app.Git.Commit(fmt.Sprintf("Start ticket: %s", t.ID)); err != nil {
		return fmt.Errorf("failed to commit ticket move: %w", err)
	}

	// Set current ticket (in main worktree)
	if err := app.Manager.SetCurrentTicket(t); err != nil {
		return fmt.Errorf("failed to set current ticket: %w", err)
	}

	// Push branch if requested
	if !noPush {
		if app.Config.Worktree.Enabled {
			// Push from worktree
			if _, err := app.Git.RunInWorktree(worktreePath, "push", "-u", "origin", t.ID); err != nil {
				fmt.Printf("Warning: Failed to push branch: %v\n", err)
			}
		} else {
			if err := app.Git.Push("origin", t.ID, true); err != nil {
				fmt.Printf("Warning: Failed to push branch: %v\n", err)
			}
		}
	}

	fmt.Printf("Started work on ticket: %s\n", t.ID)
	if app.Config.Worktree.Enabled {
		fmt.Printf("Created worktree: %s\n", worktreePath)
		fmt.Printf("\nTo work on this ticket:\n")
		fmt.Printf("  cd %s\n", worktreePath)
	} else {
		fmt.Printf("Switched to branch: %s\n", t.ID)
	}
	fmt.Printf("Created symlink: current-ticket.md -> %s\n", t.Path)

	return nil
}

// CloseTicket closes the current ticket
func (app *App) CloseTicket(noPush, force bool) error {
	// Get current ticket
	current, err := app.Manager.GetCurrentTicket()
	if err != nil {
		return err
	}
	if current == nil {
		return NewError(ErrTicketNotStarted, "No active ticket",
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
		wt, err := app.Git.FindWorktreeByBranch(current.ID)
		if err != nil {
			return fmt.Errorf("failed to find worktree: %w", err)
		}
		if wt != nil {
			isWorktree = true
			worktreePath = wt.Path

			// Check for uncommitted changes in worktree
			if !force {
				wtGit := git.New(worktreePath)
				dirty, err := wtGit.HasUncommittedChanges()
				if err != nil {
					return fmt.Errorf("failed to check worktree status: %w", err)
				}
				if dirty {
					return NewError(ErrGitDirtyWorkspace, "Uncommitted changes in worktree",
						fmt.Sprintf("Please commit your changes in %s before closing the ticket", worktreePath),
						[]string{
							fmt.Sprintf("cd %s && git commit -am 'Your message'", worktreePath),
							"Force close without committing: ticketflow close --force",
						})
				}
			}

			// Push from worktree if requested
			if !noPush {
				if _, err := app.Git.RunInWorktree(worktreePath, "push", "origin", current.ID); err != nil {
					fmt.Printf("Warning: Failed to push branch from worktree: %v\n", err)
				}
			}
		}
	}

	if !isWorktree {
		// Original behavior for non-worktree mode
		// Check for uncommitted changes
		if !force {
			dirty, err := app.Git.HasUncommittedChanges()
			if err != nil {
				return fmt.Errorf("failed to check git status: %w", err)
			}
			if dirty {
				return NewError(ErrGitDirtyWorkspace, "Uncommitted changes detected",
					"Please commit your changes before closing the ticket",
					[]string{
						"Commit your changes: git commit -am 'Your message'",
						"Force close without committing: ticketflow close --force",
					})
			}
		}

		// Get current branch
		currentBranch, err := app.Git.CurrentBranch()
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}

		// Ensure we're on the ticket branch
		if currentBranch != current.ID {
			return fmt.Errorf("not on ticket branch, expected %s but on %s", current.ID, currentBranch)
		}

		// Push if requested
		if !noPush {
			if err := app.Git.Push("origin", current.ID, false); err != nil {
				fmt.Printf("Warning: Failed to push branch: %v\n", err)
			}
		}
	}

	// Ensure we're on the default branch for merging
	currentBranch, _ := app.Git.CurrentBranch()
	if currentBranch != app.Config.Git.DefaultBranch {
		if err := app.Git.Checkout(app.Config.Git.DefaultBranch); err != nil {
			return fmt.Errorf("failed to checkout default branch: %w", err)
		}
	}

	// Squash merge
	if err := app.Git.MergeSquash(current.ID); err != nil {
		// Try to recover
		if !isWorktree {
			app.Git.Checkout(current.ID)
		}
		return fmt.Errorf("failed to merge ticket branch: %w", err)
	}

	// Update ticket status
	if err := current.Close(); err != nil {
		return err
	}

	// Move ticket file from doing to done
	oldPath := current.Path
	donePath := app.Config.GetDonePath(app.ProjectRoot)
	newPath := filepath.Join(donePath, filepath.Base(current.Path))
	
	// Move the file first
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to move ticket to done: %w", err)
	}
	
	// Update ticket data with new path
	current.Path = newPath
	if err := app.Manager.Update(current); err != nil {
		// Rollback file move
		os.Rename(newPath, oldPath)
		return fmt.Errorf("failed to update ticket: %w", err)
	}
	
	// Git add both old and new paths
	if err := app.Git.Add(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to stage ticket move: %w", err)
	}
	
	// Commit the move
	if err := app.Git.Commit(fmt.Sprintf("Close ticket: %s", current.ID)); err != nil {
		return fmt.Errorf("failed to commit ticket move: %w", err)
	}

	// Remove current ticket link
	if err := app.Manager.SetCurrentTicket(nil); err != nil {
		return fmt.Errorf("failed to remove current ticket link: %w", err)
	}

	// Remove worktree if configured
	if isWorktree && app.Config.Worktree.AutoOperations.RemoveOnClose {
		fmt.Printf("Removing worktree: %s\n", worktreePath)
		if err := app.Git.RemoveWorktree(worktreePath); err != nil {
			fmt.Printf("Warning: Failed to remove worktree: %v\n", err)
		}
	}

	fmt.Printf("Closed ticket: %s\n", current.ID)
	fmt.Printf("Merged into: %s\n", app.Config.Git.DefaultBranch)
	if isWorktree && app.Config.Worktree.AutoOperations.RemoveOnClose {
		fmt.Printf("Removed worktree: %s\n", worktreePath)
	}
	fmt.Println("Ready to commit the squashed changes")

	return nil
}

// RestoreCurrentTicket restores the current ticket symlink
func (app *App) RestoreCurrentTicket() error {
	// Get current branch
	branch, err := app.Git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Try to get ticket by branch name
	t, err := app.Manager.Get(branch)
	if err != nil {
		return fmt.Errorf("no ticket found for branch %s", branch)
	}

	// Set current ticket
	if err := app.Manager.SetCurrentTicket(t); err != nil {
		return fmt.Errorf("failed to set current ticket: %w", err)
	}

	fmt.Printf("Restored current ticket link: %s\n", t.ID)
	return nil
}

// Status shows the current status
func (app *App) Status(format OutputFormat) error {
	// Get current ticket
	current, err := app.Manager.GetCurrentTicket()
	if err != nil {
		return err
	}

	// Get current branch
	branch, err := app.Git.CurrentBranch()
	if err != nil {
		return err
	}

	// Get ticket stats
	allTickets, err := app.Manager.List("")
	if err != nil {
		return err
	}

	todoCount := 0
	doingCount := 0
	doneCount := 0

	for _, t := range allTickets {
		switch t.Status() {
		case ticket.StatusTodo:
			todoCount++
		case ticket.StatusDoing:
			doingCount++
		case ticket.StatusDone:
			doneCount++
		}
	}

	if format == FormatJSON {
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

	// Text format
	fmt.Printf("Current branch: %s\n", branch)

	if current != nil {
		fmt.Printf("\nCurrent ticket: %s\n", current.ID)
		fmt.Printf("  Description: %s\n", current.Description)
		fmt.Printf("  Status: %s\n", current.Status())
		if current.StartedAt != nil {
			duration := time.Since(*current.StartedAt)
			fmt.Printf("  Duration: %s\n", formatDuration(duration))
		}
	} else {
		fmt.Println("\nNo active ticket")
	}

	fmt.Printf("\nTicket summary:\n")
	fmt.Printf("  Todo:  %d\n", todoCount)
	fmt.Printf("  Doing: %d\n", doingCount)
	fmt.Printf("  Done:  %d\n", doneCount)
	fmt.Printf("  Total: %d\n", len(allTickets))

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
	if strings.Contains(content, "current-ticket.md") {
		return nil
	}

	// Append our entries
	toAdd := "\n# TicketFlow\ncurrent-ticket.md\n.worktrees/\n"

	// Write back
	return os.WriteFile(path, []byte(content+toAdd), 0644)
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

func (app *App) outputTicketListJSON(tickets []*ticket.Ticket) error {
	ticketList := make([]map[string]interface{}, len(tickets))
	for i, t := range tickets {
		ticketList[i] = ticketToJSON(t, "")
	}

	// Calculate summary
	todoCount := 0
	doingCount := 0
	doneCount := 0

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

	output := map[string]interface{}{
		"tickets": ticketList,
		"summary": map[string]int{
			"total": len(tickets),
			"todo":  todoCount,
			"doing": doingCount,
			"done":  doneCount,
		},
	}

	return outputJSON(output)
}

// ListWorktrees lists all worktrees
func (app *App) ListWorktrees(format OutputFormat) error {
	worktrees, err := app.Git.ListWorktrees()
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
func (app *App) CleanWorktrees() error {
	// First prune to clean up git's internal state
	if err := app.Git.PruneWorktrees(); err != nil {
		return fmt.Errorf("failed to prune worktrees: %w", err)
	}

	// Get all worktrees
	worktrees, err := app.Git.ListWorktrees()
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Get all active tickets
	activeTickets, err := app.Manager.List(string(ticket.StatusDoing))
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
			if err := app.Git.RemoveWorktree(wt.Path); err != nil {
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