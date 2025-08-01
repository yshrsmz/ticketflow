package cli

import (
	"errors"
	"fmt"

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
			"Start the ticket first with 'ticketflow start <ticket-id>'",
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
		return NewError(ErrGitMergeFailed, fmt.Sprintf("Git operation failed: %s", gitErr.Op), err.Error(), nil)
	}

	var worktreeErr *ticketerrors.WorktreeError
	if errors.As(err, &worktreeErr) {
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
