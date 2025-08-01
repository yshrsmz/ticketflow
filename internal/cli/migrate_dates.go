package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yshrsmz/ticketflow/internal/log"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// rfc3339NanoRegex matches RFC3339Nano timestamps (with subseconds)
var rfc3339NanoRegex = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+`)

// MigrateDates updates all ticket files to use standardized date format
func (app *App) MigrateDates(ctx context.Context, dryRun bool) error {
	logger := log.Global().WithOperation("migrate_dates")
	logger.Info("starting date migration", "dry_run", dryRun)

	// Get all tickets
	tickets, err := app.Manager.List(ctx, ticket.StatusFilterAll)
	if err != nil {
		logger.WithError(err).Error("failed to list tickets")
		return fmt.Errorf("failed to list tickets: %w", err)
	}
	logger.Debug("found tickets", "count", len(tickets))

	updatedCount := 0
	for _, t := range tickets {
		// Read the original file
		originalContent, err := os.ReadFile(t.Path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", t.Path, err)
		}

		// Check if the file contains microseconds in dates
		// Look for patterns like: 2025-07-26T18:14:10.48619+09:00
		hasSubseconds := false
		contentStr := string(originalContent)

		// Check each date field for subseconds
		for _, field := range []string{"created_at:", "started_at:", "closed_at:"} {
			if idx := strings.Index(contentStr, field); idx != -1 {
				// Extract the line containing the date field
				lineEnd := strings.IndexByte(contentStr[idx:], '\n')
				if lineEnd == -1 {
					lineEnd = len(contentStr) - idx
				}
				line := contentStr[idx : idx+lineEnd]
				// Check if the line matches the RFC3339Nano regex
				if rfc3339NanoRegex.MatchString(line) {
					hasSubseconds = true
					break
				}
			}
		}

		if hasSubseconds {
			// Parse and re-save the ticket to apply new formatting
			parsedTicket, err := ticket.Parse(originalContent)
			if err != nil {
				logger.WithError(err).Warn("failed to parse ticket", "path", t.Path)
				fmt.Printf("Warning: failed to parse %s: %v\n", t.Path, err)
				continue
			}

			// Copy metadata from the loaded ticket
			parsedTicket.ID = t.ID
			parsedTicket.Slug = t.Slug
			parsedTicket.Path = t.Path

			if dryRun {
				logger.Debug("would update ticket", "path", t.Path)
				fmt.Printf("Would update: %s\n", filepath.Base(t.Path))
			} else {
				// Write back with new format
				data, err := parsedTicket.ToBytes()
				if err != nil {
					logger.WithError(err).Error("failed to serialize ticket", "path", t.Path)
					return fmt.Errorf("failed to serialize %s: %w", t.Path, err)
				}

				if err := os.WriteFile(t.Path, data, 0644); err != nil {
					logger.WithError(err).Error("failed to write ticket", "path", t.Path)
					return fmt.Errorf("failed to write %s: %w", t.Path, err)
				}

				logger.Info("updated ticket", "path", t.Path)
				fmt.Printf("Updated: %s\n", filepath.Base(t.Path))
			}
			updatedCount++
		}
	}

	logger.Info("migration completed", "updated_count", updatedCount, "dry_run", dryRun)
	if dryRun {
		fmt.Printf("\nDry run complete. Would update %d ticket(s)\n", updatedCount)
	} else {
		fmt.Printf("\nMigration complete. Updated %d ticket(s)\n", updatedCount)
	}

	return nil
}
