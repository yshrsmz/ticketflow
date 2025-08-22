// Package commands provides CLI command implementations and utilities.
// This file contains validation and extraction helper functions used across commands
// to reduce code duplication and maintain consistency.
package commands

import (
	"fmt"
	"strings"

	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// Error message constants for consistent error formatting
const (
	// ErrInvalidFormat is the error message template for invalid format values
	ErrInvalidFormat = "invalid format: %q (must be %q or %q)"
	// ErrInvalidFlags is the error message template for invalid flag type assertions
	ErrInvalidFlags = "invalid flags type: expected *%T, got %T"
)

// ValidateFormat validates that the format string is either "text" or "json".
// Returns an error if the format is invalid.
func ValidateFormat(format string) error {
	if format != FormatText && format != FormatJSON {
		return fmt.Errorf(ErrInvalidFormat, format, FormatText, FormatJSON)
	}
	return nil
}

// ExtractParentFromTicket extracts the parent ticket ID from a ticket's related field.
// Returns an empty string if the ticket is nil, has no related items, or has no parent.
func ExtractParentFromTicket(t *ticket.Ticket) string {
	if t == nil || len(t.Related) == 0 {
		return ""
	}
	for _, rel := range t.Related {
		if strings.HasPrefix(rel, "parent:") {
			return strings.TrimPrefix(rel, "parent:")
		}
	}
	return ""
}

// AssertFlags performs a type assertion on the flags interface to the specified type T.
// Returns an error if the type assertion fails, providing helpful error messages.
func AssertFlags[T any](flags interface{}) (*T, error) {
	f, ok := flags.(*T)
	if !ok {
		return nil, fmt.Errorf(ErrInvalidFlags, *new(T), flags)
	}
	return f, nil
}
