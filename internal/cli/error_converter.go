package cli

import (
	"errors"
	"fmt"
	"strings"

	ticketerrors "github.com/yshrsmz/ticketflow/internal/errors"
)

// ConvertError converts internal errors to CLI errors with appropriate codes and suggestions
func ConvertError(err error) error {
	if err == nil {
		return nil
	}

	// Check if it's already a CLI error
	if _, ok := err.(*CLIError); ok {
		return err
	}

	// Handle sentinel errors
	switch {
	case errors.Is(err, ticketerrors.ErrTicketNotFound):
		return NewError(ErrTicketNotFound, "Ticket not found", err.Error(), []string{
			"Check the ticket ID is correct",
			"Use 'ticketflow list' to see available tickets",
		})

	case errors.Is(err, ticketerrors.ErrTicketExists):
		return NewError(ErrTicketExists, "Ticket already exists", err.Error(), []string{
			"Use a different ticket ID",
			"Check existing tickets with 'ticketflow list'",
		})

	case errors.Is(err, ticketerrors.ErrTicketAlreadyStarted):
		return NewError(ErrTicketAlreadyStarted, "Ticket already started", err.Error(), []string{
			"This ticket is already in progress",
			"Use 'ticketflow status' to see current ticket",
		})

	case errors.Is(err, ticketerrors.ErrTicketAlreadyClosed):
		return NewError(ErrTicketAlreadyClosed, "Ticket already closed", err.Error(), []string{
			"This ticket has already been completed",
			"Create a new ticket if you need to continue work",
		})

	case errors.Is(err, ticketerrors.ErrTicketNotStarted):
		return NewError(ErrTicketNotStarted, "Ticket not started", err.Error(), []string{
			"If in a worktree, try restoring the current ticket link: 'ticketflow restore'",
			"Or start the ticket first with 'ticketflow start <ticket-id>'",
		})

	case errors.Is(err, ticketerrors.ErrNotGitRepo):
		return NewError(ErrNotGitRepo, "Not in a git repository", err.Error(), []string{
			"Change to a git repository directory",
			"Initialize a new git repository with 'git init'",
		})

	case errors.Is(err, ticketerrors.ErrWorktreeExists):
		return NewError(ErrWorktreeExists, "Worktree already exists", err.Error(), []string{
			"The worktree for this ticket already exists",
			"Use 'ticketflow worktree list' to see existing worktrees",
		})

	case errors.Is(err, ticketerrors.ErrWorktreeNotFound):
		return NewError(ErrWorktreeNotFound, "Worktree not found", err.Error(), []string{
			"The worktree for this ticket doesn't exist",
			"Start the ticket with 'ticketflow start <ticket-id>' to create a worktree",
		})

	case errors.Is(err, ticketerrors.ErrConfigNotFound):
		return NewError(ErrConfigNotFound, "Configuration not found", err.Error(), []string{
			"Run 'ticketflow init' to initialize the ticket system",
		})

	case errors.Is(err, ticketerrors.ErrConfigInvalid):
		return NewError(ErrConfigInvalid, "Invalid configuration", err.Error(), []string{
			"Check your .ticketflow.yaml file for errors",
			"Run 'ticketflow init' to recreate the configuration",
		})
	}

	// Handle typed errors
	var ticketErr *ticketerrors.TicketError
	if errors.As(err, &ticketErr) {
		return NewError(ErrTicketInvalid, fmt.Sprintf("Ticket operation failed: %s", ticketErr.Op), err.Error(), nil)
	}

	var gitErr *ticketerrors.GitError
	if errors.As(err, &gitErr) {
		// Check for worktree-specific git errors
		if gitErr.Op == "worktree" || (gitErr.Err != nil && strings.Contains(gitErr.Err.Error(), "worktree")) {
			if enhanced := enhanceWorktreeGitError(gitErr); enhanced != nil {
				return enhanced
			}
		}
		return NewError(ErrGitMergeFailed, fmt.Sprintf("Git operation failed: %s", gitErr.Op), err.Error(), nil)
	}

	var worktreeErr *ticketerrors.WorktreeError
	if errors.As(err, &worktreeErr) {
		// Try to enhance worktree error first
		if enhanced := enhanceWorktreeError(worktreeErr); enhanced != nil {
			return enhanced
		}

		code := ErrWorktreeCreateFailed
		if worktreeErr.Op == "remove" {
			code = ErrWorktreeRemoveFailed
		}
		return NewError(code, fmt.Sprintf("Worktree operation failed: %s", worktreeErr.Op), err.Error(), nil)
	}

	var configErr *ticketerrors.ConfigError
	if errors.As(err, &configErr) {
		return NewError(ErrConfigInvalid, "Configuration error", err.Error(), []string{
			fmt.Sprintf("Check the '%s' field in your .ticketflow.yaml", configErr.Field),
		})
	}

	// Generic error
	return err
}

// enhanceWorktreeGitError enhances git errors related to worktree operations
func enhanceWorktreeGitError(gitErr *ticketerrors.GitError) *CLIError {
	if gitErr == nil || gitErr.Err == nil {
		return nil
	}

	return enhanceWorktreeErrorString(gitErr.Err.Error(), gitErr.Error())
}

// enhanceWorktreeError enhances worktree-specific errors
func enhanceWorktreeError(worktreeErr *ticketerrors.WorktreeError) *CLIError {
	if worktreeErr == nil || worktreeErr.Err == nil {
		return nil
	}

	return enhanceWorktreeErrorString(worktreeErr.Err.Error(), worktreeErr.Error())
}

// enhanceWorktreeErrorString provides common enhancement logic for worktree error messages
func enhanceWorktreeErrorString(errStr, fullError string) *CLIError {
	// Pattern matching for common worktree errors
	switch {
	case strings.Contains(errStr, "is not a working tree"):
		return NewError(ErrWorktreeRemoveFailed, "Worktree appears to be corrupted",
			fullError,
			[]string{
				"Run 'git worktree prune' to clean up corrupted references",
				"Then retry your command",
			})

	case strings.Contains(errStr, "already exists"):
		return NewError(ErrWorktreeExists, "Worktree directory already exists",
			fullError,
			[]string{
				"Remove the directory manually if it's no longer needed",
				"Or use 'ticketflow cleanup' if it's an old ticket",
			})

	case strings.Contains(errStr, "is already checked out"):
		return NewError(ErrWorktreeExists, "Branch is already checked out in another worktree",
			fullError,
			[]string{
				"Use 'git worktree list' to find where it's checked out",
				"Remove the other worktree if it's no longer needed",
			})

	case strings.Contains(errStr, "could not create work tree dir"):
		return NewError(ErrWorktreeCreateFailed, "Cannot create worktree directory",
			fullError,
			[]string{
				"Check directory permissions",
				"Ensure you have enough disk space",
				"Try 'git worktree prune' if references are corrupted",
			})

	case strings.Contains(errStr, "invalid reference"):
		return NewError(ErrWorktreeCreateFailed, "Invalid git reference for worktree",
			fullError,
			[]string{
				"Ensure the branch name is valid",
				"Check if the base branch exists",
				"Try 'git fetch' to update remote references",
			})

	case strings.Contains(errStr, "permission denied"):
		return NewError(ErrPermissionDenied, "Permission denied for worktree operation",
			fullError,
			[]string{
				"Check file and directory permissions",
				"Ensure you have write access to the parent directory",
				"Try running with appropriate permissions",
			})

	case strings.Contains(errStr, "locked"):
		return NewError(ErrWorktreeRemoveFailed, "Worktree is locked",
			fullError,
			[]string{
				"Check if another process is using the worktree",
				"Remove lock file manually if stale: .git/worktrees/<name>/locked",
				"Use 'git worktree remove --force' if necessary",
			})
	}

	return nil // No enhancement needed
}
