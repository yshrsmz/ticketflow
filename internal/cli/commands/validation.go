// Package commands provides CLI command implementations and utilities.
// This file contains validation and extraction helper functions used across commands
// to reduce code duplication and maintain consistency.
package commands

import (
	"fmt"
	"strings"

	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// ValidateFormat validates that the format string is either "text" or "json".
// Returns an error if the format is invalid.
func ValidateFormat(format string) error {
	if format != FormatText && format != FormatJSON {
		return fmt.Errorf("invalid format: %q (must be %q or %q)", format, FormatText, FormatJSON)
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
		return nil, fmt.Errorf("invalid flags type: expected *%T, got %T", *new(T), flags)
	}
	return f, nil
}