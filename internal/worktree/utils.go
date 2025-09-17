package worktree

import (
	"context"
	"path/filepath"

	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/log"
)

// GetPath attempts to get the actual worktree path for a ticket, or falls back to calculating it.
// The repoRoot should point to the primary repository root (not a linked worktree) so calculated
// paths are anchored correctly even when called from within a worktree.
//
// The fallback scenario occurs when:
// - The worktree was created outside of ticketflow
// - The git worktree state is inconsistent
// - The worktree reference exists but cannot be queried
// - During error handling when we know a worktree exists but can't get its details
func GetPath(ctx context.Context, gitClient git.GitClient, cfg *config.Config, repoRoot, ticketID string) string {
	logger := log.Global().WithTicket(ticketID)

	// Try to get the actual worktree path from git
	wt, err := gitClient.FindWorktreeByBranch(ctx, ticketID)
	if err != nil {
		// Log the error for debugging purposes
		logger.Debug("failed to find worktree by branch", "error", err, "ticketID", ticketID)
	} else if wt != nil {
		logger.Debug("found worktree path from git", "path", wt.Path)
		return wt.Path
	}

	// Fall back to calculated path based on configuration
	// This ensures we always return a valid path even if git state is inconsistent
	baseDir := cfg.GetWorktreePath(repoRoot)
	calculatedPath := filepath.Join(baseDir, ticketID)
	logger.Debug("using calculated worktree path", "path", calculatedPath, "reason", "worktree not found in git")
	return calculatedPath
}
