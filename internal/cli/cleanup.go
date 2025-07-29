package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// AutoCleanup performs automatic cleanup of old tickets and worktrees
func (app *App) AutoCleanup(dryRun bool) error {
	fmt.Println("Starting auto-cleanup...")

	// 1. Clean orphaned worktrees
	if app.Config.Worktree.Enabled {
		if err := app.cleanOrphanedWorktrees(dryRun); err != nil {
			fmt.Printf("Warning: Failed to clean worktrees: %v\n", err)
		}
	}

	// 2. Clean up old done tickets (optional future enhancement)
	// For now, done tickets stay in done/ directory permanently

	// 3. Clean up stale branches (done tickets without worktrees)
	if err := app.cleanStaleBranches(dryRun); err != nil {
		fmt.Printf("Warning: Failed to clean branches: %v\n", err)
	}

	fmt.Println("Auto-cleanup completed.")
	return nil
}

// cleanOrphanedWorktrees removes worktrees without active tickets
func (app *App) cleanOrphanedWorktrees(dryRun bool) error {
	if !app.Config.Worktree.Enabled {
		return nil
	}

	fmt.Println("\nCleaning orphaned worktrees...")

	// First prune to clean up git's internal state
	if !dryRun {
		if err := app.Git.PruneWorktrees(); err != nil {
			return fmt.Errorf("failed to prune worktrees: %w", err)
		}
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
			fmt.Printf("  Removing orphaned worktree: %s (branch: %s)\n", wt.Path, wt.Branch)

			if !dryRun {
				if err := app.Git.RemoveWorktree(wt.Path); err != nil {
					fmt.Printf("  Warning: Failed to remove worktree %s: %v\n", wt.Path, err)
				} else {
					cleaned++
				}
			} else {
				cleaned++
			}
		}
	}

	fmt.Printf("  Cleaned %d orphaned worktree(s)\n", cleaned)
	return nil
}

// cleanStaleBranches removes branches for done tickets
func (app *App) cleanStaleBranches(dryRun bool) error {
	fmt.Println("\nCleaning stale branches...")

	// Get all branches
	output, err := app.Git.Exec("branch", "--format=%(refname:short)")
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	branches := splitLines(output)

	// Get all tickets (including done ones)
	// Pass "all" to include done tickets
	allTickets, err := app.Manager.List("all")
	if err != nil {
		return fmt.Errorf("failed to list tickets: %w", err)
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
				fmt.Printf("  Removing branch for done ticket: %s\n", branch)

				if !dryRun {
					// Delete local branch (force delete to avoid warnings)
					if _, err := app.Git.Exec("branch", "-D", branch); err != nil {
						fmt.Printf("  Warning: Failed to delete branch %s: %v\n", branch, err)
					} else {
						cleaned++
					}
				} else {
					cleaned++
				}
			}
		}
	}

	fmt.Printf("  Cleaned %d stale branch(es)\n", cleaned)
	return nil
}

// CleanupStats shows what would be cleaned up
func (app *App) CleanupStats() error {
	fmt.Println("Cleanup Statistics:")
	fmt.Println("==================")

	// Done tickets statistics
	doneTickets, err := app.Manager.List(string(ticket.StatusDone))
	if err == nil {
		fmt.Printf("\nDone tickets: %d\n", len(doneTickets))
	}

	// Worktree statistics
	if app.Config.Worktree.Enabled {
		worktrees, err := app.Git.ListWorktrees()
		activeTickets, _ := app.Manager.List(string(ticket.StatusDoing))

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
			fmt.Printf("Orphaned worktrees: %d\n", orphaned)
		}
	}

	// Branch statistics
	output, err := app.Git.Exec("branch", "--format=%(refname:short)")
	if err == nil {
		branches := splitLines(output)
		allTickets, _ := app.Manager.List("all")

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
		fmt.Printf("Stale branches (done tickets): %d\n", stale)
	}

	// Check done directory size
	donePath := app.Config.GetDonePath(app.ProjectRoot)
	if info, err := os.Stat(donePath); err == nil && info.IsDir() {
		entries, _ := os.ReadDir(donePath)
		fmt.Printf("Done directory tickets: %d\n", len(entries))
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
