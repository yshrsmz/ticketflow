package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/yshrsmz/ticketflow/internal/log"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// CleanupResult holds the statistics from cleanup operations
type CleanupResult struct {
	OrphanedWorktrees int
	StaleBranches     int
	Errors            []string
}

// HasErrors returns true if any errors occurred during cleanup
func (r *CleanupResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// AutoCleanup performs automatic cleanup of old tickets and worktrees
func (app *App) AutoCleanup(ctx context.Context, dryRun bool) (*CleanupResult, error) {
	logger := log.Global().WithOperation("auto_cleanup")
	logger.Info("starting auto-cleanup", slog.Bool("dry_run", dryRun))

	// Ensure StatusWriter is initialized (for tests that don't set it)
	if app.StatusWriter == nil {
		app.StatusWriter = NewNullStatusWriter()
	}

	app.StatusWriter.Println("Starting auto-cleanup...")

	result := &CleanupResult{
		Errors: make([]string, 0),
	}

	// 1. Clean orphaned worktrees
	if app.Config.Worktree.Enabled {
		cleaned, err := app.cleanOrphanedWorktrees(ctx, dryRun)
		if err != nil {
			logger.WithError(err).Warn("failed to clean worktrees")
			app.StatusWriter.Printf("Warning: Failed to clean worktrees: %v\n", err)
			result.Errors = append(result.Errors, fmt.Sprintf("worktrees: %v", err))
		} else {
			result.OrphanedWorktrees = cleaned
			logger.Info("cleaned orphaned worktrees", "count", cleaned)
		}
	}

	// 2. Clean up old done tickets (optional future enhancement)
	// For now, done tickets stay in done/ directory permanently

	// 3. Clean up stale branches (done tickets without worktrees)
	cleaned, err := app.cleanStaleBranches(ctx, dryRun)
	if err != nil {
		logger.WithError(err).Warn("failed to clean branches")
		app.StatusWriter.Printf("Warning: Failed to clean branches: %v\n", err)
		result.Errors = append(result.Errors, fmt.Sprintf("branches: %v", err))
	} else {
		result.StaleBranches = cleaned
		logger.Info("cleaned stale branches", "count", cleaned)
	}

	logger.Info("auto-cleanup completed", "orphaned_worktrees", result.OrphanedWorktrees, "stale_branches", result.StaleBranches, "errors", len(result.Errors))
	app.StatusWriter.Println("Auto-cleanup completed.")
	return result, nil
}

// cleanOrphanedWorktrees removes worktrees without active tickets
func (app *App) cleanOrphanedWorktrees(ctx context.Context, dryRun bool) (int, error) {
	logger := log.Global().WithOperation("clean_orphaned_worktrees")

	if !app.Config.Worktree.Enabled {
		return 0, nil
	}

	// Ensure StatusWriter is initialized (for tests that don't set it)
	if app.StatusWriter == nil {
		app.StatusWriter = NewNullStatusWriter()
	}

	logger.Debug("cleaning orphaned worktrees", "dry_run", dryRun)
	app.StatusWriter.Println("\nCleaning orphaned worktrees...")

	// First prune to clean up git's internal state
	if !dryRun {
		if err := app.Git.PruneWorktrees(ctx); err != nil {
			logger.WithError(err).Error("failed to prune worktrees")
			return 0, fmt.Errorf("failed to prune worktrees: %w", err)
		}
		logger.Debug("pruned worktrees")
	}

	// Get all worktrees
	worktrees, err := app.Git.ListWorktrees(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Get all active tickets
	activeTickets, err := app.Manager.List(ctx, ticket.StatusFilterDoing)
	if err != nil {
		return 0, fmt.Errorf("failed to list active tickets: %w", err)
	}

	// Create map of active ticket IDs
	activeMap := make(map[string]bool)
	for _, t := range activeTickets {
		activeMap[t.ID] = true
	}

	cleaned := 0
	for _, wt := range worktrees {
		// Skip main worktree
		if wt.Branch == "" || wt.Branch == app.Config.Git.DefaultBranch {
			continue
		}

		// Check if this worktree has an active ticket
		if !activeMap[wt.Branch] {
			logger.Info("removing orphaned worktree", "path", wt.Path, "branch", wt.Branch)
			app.StatusWriter.Printf("  Removing orphaned worktree: %s (branch: %s)\n", wt.Path, wt.Branch)

			if !dryRun {
				if err := app.Git.RemoveWorktree(ctx, wt.Path); err != nil {
					logger.WithError(err).Warn("failed to remove worktree", "path", wt.Path)
					app.StatusWriter.Printf("  Warning: Failed to remove worktree %s: %v\n", wt.Path, err)
				} else {
					cleaned++
				}
			} else {
				cleaned++
			}
		}
	}

	logger.Info("cleaned orphaned worktrees", "count", cleaned)
	app.StatusWriter.Printf("  Cleaned %d orphaned worktree(s)\n", cleaned)
	return cleaned, nil
}

// cleanStaleBranches removes branches for done tickets
func (app *App) cleanStaleBranches(ctx context.Context, dryRun bool) (int, error) {
	logger := log.Global().WithOperation("clean_stale_branches")

	// Ensure StatusWriter is initialized (for tests that don't set it)
	if app.StatusWriter == nil {
		app.StatusWriter = NewNullStatusWriter()
	}

	logger.Debug("cleaning stale branches", "dry_run", dryRun)
	app.StatusWriter.Println("\nCleaning stale branches...")

	// Get all branches
	output, err := app.Git.Exec(ctx, "branch", "--format=%(refname:short)")
	if err != nil {
		return 0, fmt.Errorf("failed to list branches: %w", err)
	}

	branches := splitLines(output)

	// Get all tickets (including done ones)
	// Pass StatusFilterAll to include done tickets
	allTickets, err := app.Manager.List(ctx, ticket.StatusFilterAll)
	if err != nil {
		return 0, fmt.Errorf("failed to list tickets: %w", err)
	}

	// Create map of ticket IDs and their status
	ticketStatus := make(map[string]ticket.Status)
	for _, t := range allTickets {
		ticketStatus[t.ID] = t.Status()
	}

	cleaned := 0
	for _, branch := range branches {
		// Skip main/master branches
		if branch == app.Config.Git.DefaultBranch || branch == "main" || branch == "master" {
			continue
		}

		// Check if this is a ticket branch
		if status, exists := ticketStatus[branch]; exists {
			// Remove branches for done tickets
			if status == ticket.StatusDone {
				logger.Info("removing branch for done ticket", "branch", branch)
				app.StatusWriter.Printf("  Removing branch for done ticket: %s\n", branch)

				if !dryRun {
					// Delete local branch (force delete to avoid warnings)
					if _, err := app.Git.Exec(ctx, "branch", "-D", branch); err != nil {
						logger.WithError(err).Warn("failed to delete branch", "branch", branch)
						app.StatusWriter.Printf("  Warning: Failed to delete branch %s: %v\n", branch, err)
					} else {
						cleaned++
					}
				} else {
					cleaned++
				}
			}
		}
	}

	logger.Info("cleaned stale branches", "count", cleaned)
	app.StatusWriter.Printf("  Cleaned %d stale branch(es)\n", cleaned)
	return cleaned, nil
}

// CleanupStats shows what would be cleaned up
func (app *App) CleanupStats(ctx context.Context) error {
	// Ensure StatusWriter is initialized (for tests that don't set it)
	if app.StatusWriter == nil {
		app.StatusWriter = NewNullStatusWriter()
	}

	app.StatusWriter.Println("Cleanup Statistics:")
	app.StatusWriter.Println("==================")

	// Done tickets statistics
	doneTickets, err := app.Manager.List(ctx, ticket.StatusFilterDone)
	if err == nil {
		app.StatusWriter.Printf("\nDone tickets: %d\n", len(doneTickets))
	}

	// Worktree statistics
	if app.Config.Worktree.Enabled {
		worktrees, err := app.Git.ListWorktrees(ctx)
		activeTickets, _ := app.Manager.List(ctx, ticket.StatusFilterDoing)

		if err == nil {
			activeMap := make(map[string]bool)
			for _, t := range activeTickets {
				activeMap[t.ID] = true
			}

			orphaned := 0
			for _, wt := range worktrees {
				if wt.Branch != "" && wt.Branch != app.Config.Git.DefaultBranch && !activeMap[wt.Branch] {
					orphaned++
				}
			}
			app.StatusWriter.Printf("Orphaned worktrees: %d\n", orphaned)
		}
	}

	// Branch statistics
	output, err := app.Git.Exec(ctx, "branch", "--format=%(refname:short)")
	if err == nil {
		branches := splitLines(output)
		allTickets, _ := app.Manager.List(ctx, ticket.StatusFilterAll)

		ticketStatus := make(map[string]ticket.Status)
		for _, t := range allTickets {
			ticketStatus[t.ID] = t.Status()
		}

		stale := 0
		for _, branch := range branches {
			if status, exists := ticketStatus[branch]; exists && status == ticket.StatusDone {
				stale++
			}
		}
		app.StatusWriter.Printf("Stale branches (done tickets): %d\n", stale)
	}

	// Check done directory size
	donePath := app.Config.GetDonePath(app.ProjectRoot)
	if info, err := os.Stat(donePath); err == nil && info.IsDir() {
		entries, _ := os.ReadDir(donePath)
		app.StatusWriter.Printf("Done directory tickets: %d\n", len(entries))
	}

	return nil
}

// splitLines splits a string into lines, filtering out empty lines
func splitLines(s string) []string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}
