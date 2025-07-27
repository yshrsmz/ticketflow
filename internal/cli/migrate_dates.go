package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// MigrateDates updates all ticket files to use standardized date format
func (app *App) MigrateDates(dryRun bool) error {
	// Get all tickets
	tickets, err := app.Manager.List("all")
	if err != nil {
		return fmt.Errorf("failed to list tickets: %w", err)
	}

	updatedCount := 0
	for _, t := range tickets {
		// Read the original file
		originalContent, err := os.ReadFile(t.Path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", t.Path, err)
		}

		// Check if the file contains microseconds in dates
		if strings.Contains(string(originalContent), ".") &&
			(strings.Contains(string(originalContent), "created_at:") ||
				strings.Contains(string(originalContent), "started_at:") ||
				strings.Contains(string(originalContent), "closed_at:")) {

			// Parse and re-save the ticket to apply new formatting
			parsedTicket, err := ticket.Parse(originalContent)
			if err != nil {
				fmt.Printf("Warning: failed to parse %s: %v\n", t.Path, err)
				continue
			}

			// Copy metadata from the loaded ticket
			parsedTicket.ID = t.ID
			parsedTicket.Slug = t.Slug
			parsedTicket.Path = t.Path

			if dryRun {
				fmt.Printf("Would update: %s\n", filepath.Base(t.Path))
			} else {
				// Write back with new format
				data, err := parsedTicket.ToBytes()
				if err != nil {
					return fmt.Errorf("failed to serialize %s: %w", t.Path, err)
				}

				if err := os.WriteFile(t.Path, data, 0644); err != nil {
					return fmt.Errorf("failed to write %s: %w", t.Path, err)
				}

				fmt.Printf("Updated: %s\n", filepath.Base(t.Path))
			}
			updatedCount++
		}
	}

	if dryRun {
		fmt.Printf("\nDry run complete. Would update %d ticket(s)\n", updatedCount)
	} else {
		fmt.Printf("\nMigration complete. Updated %d ticket(s)\n", updatedCount)
	}

	return nil
}

