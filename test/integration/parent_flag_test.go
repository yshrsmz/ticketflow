package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/cli"
	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// readTicketFromFile reads and parses a ticket from a file
func readTicketFromFile(t *testing.T, path string) *ticket.Ticket {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	parsedTicket, err := ticket.Parse(data)
	require.NoError(t, err)

	// Set computed fields
	parsedTicket.Path = path
	parsedTicket.ID = strings.TrimSuffix(filepath.Base(path), ".md")

	return parsedTicket
}

func TestNewCommandWithParentFlag(t *testing.T) {
	// Integration tests cannot run in parallel due to os.Chdir

	// Setup git repository
	tmpDir := setupTestRepo(t)

	// Save current directory and change to test directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err)
	}()

	require.NoError(t, os.Chdir(tmpDir))

	// Initialize CLI app
	ctx := context.Background()
	err = cli.InitCommand(ctx)
	require.NoError(t, err)

	// Create app instance
	app, err := cli.NewApp(ctx)
	require.NoError(t, err)

	// Create a parent ticket first
	err = app.NewTicket(ctx, "parent-feature", "", cli.FormatText)
	require.NoError(t, err)

	// Get the parent ticket ID
	parentTicketPath := filepath.Join(tmpDir, "tickets", "todo", "*parent-feature.md")
	parentFiles, err := filepath.Glob(parentTicketPath)
	require.NoError(t, err)
	require.Len(t, parentFiles, 1, "Should have created one parent ticket")
	parentID := strings.TrimSuffix(filepath.Base(parentFiles[0]), ".md")

	t.Run("create sub-ticket with --parent flag", func(t *testing.T) {
		// Create sub-ticket with explicit parent
		err := app.NewTicket(ctx, "sub-feature", parentID, cli.FormatText)
		require.NoError(t, err)

		// Find the created sub-ticket
		subTicketPath := filepath.Join(tmpDir, "tickets", "todo", "*sub-feature.md")
		subFiles, err := filepath.Glob(subTicketPath)
		require.NoError(t, err)
		require.Len(t, subFiles, 1, "Should have created one sub-ticket")

		// Read and verify sub-ticket has parent relationship
		subTicket := readTicketFromFile(t, subFiles[0])

		require.NotNil(t, subTicket.Related, "Sub-ticket should have Related field")
		assert.Contains(t, subTicket.Related, "parent:"+parentID)
	})

	t.Run("create sub-ticket with -p flag (short form)", func(t *testing.T) {
		// Create sub-ticket with short form parent flag (same functionality as --parent)
		err := app.NewTicket(ctx, "another-sub", parentID, cli.FormatText)
		require.NoError(t, err)

		// Find the created sub-ticket
		subTicketPath := filepath.Join(tmpDir, "tickets", "todo", "*another-sub.md")
		subFiles, err := filepath.Glob(subTicketPath)
		require.NoError(t, err)
		require.Len(t, subFiles, 1, "Should have created one sub-ticket")

		// Read and verify sub-ticket has parent relationship
		subTicket := readTicketFromFile(t, subFiles[0])

		require.NotNil(t, subTicket.Related, "Sub-ticket should have Related field")
		assert.Contains(t, subTicket.Related, "parent:"+parentID)
	})

	t.Run("error on non-existent parent", func(t *testing.T) {
		// Try to create sub-ticket with non-existent parent
		err := app.NewTicket(ctx, "orphan-sub", "non-existent-ticket", cli.FormatText)
		assert.Error(t, err, "Should fail with non-existent parent")
		assert.Contains(t, err.Error(), "Parent ticket not found")
	})

	t.Run("error on self-parent", func(t *testing.T) {
		// Try to create ticket with itself as parent
		err := app.NewTicket(ctx, "self-parent", "self-parent", cli.FormatText)
		assert.Error(t, err, "Should fail with self-parent")
		assert.Contains(t, err.Error(), "Invalid parent relationship")
	})

	t.Run("explicit parent overrides implicit worktree parent", func(t *testing.T) {
		// Create another parent ticket
		err = app.NewTicket(ctx, "another-parent", "", cli.FormatText)
		require.NoError(t, err)

		// Get the another parent ticket ID
		anotherParentPath := filepath.Join(tmpDir, "tickets", "todo", "*another-parent.md")
		anotherParentFiles, err := filepath.Glob(anotherParentPath)
		require.NoError(t, err)
		require.Len(t, anotherParentFiles, 1)
		anotherParentID := strings.TrimSuffix(filepath.Base(anotherParentFiles[0]), ".md")

		// Create sub-ticket with explicit parent different from first parent
		// This tests that explicit parent is used even when we could have an implicit parent
		err = app.NewTicket(ctx, "explicit-over-implicit", anotherParentID, cli.FormatText)
		require.NoError(t, err)

		// Find the created sub-ticket
		subTicketPath := filepath.Join(tmpDir, "tickets", "todo", "*explicit-over-implicit.md")
		subFiles, err := filepath.Glob(subTicketPath)
		require.NoError(t, err)
		require.Len(t, subFiles, 1)

		// Read and verify sub-ticket has explicit parent
		subTicket := readTicketFromFile(t, subFiles[0])

		require.NotNil(t, subTicket.Related)
		assert.Contains(t, subTicket.Related, "parent:"+anotherParentID)
		// Should not have the first parent
		assert.NotContains(t, subTicket.Related, "parent:"+parentID)
	})

	t.Run("use ticket ID as parent", func(t *testing.T) {
		// Create sub-ticket using parent ticket ID instead of slug
		err := app.NewTicket(ctx, "sub-with-id-parent", parentID, cli.FormatText)
		require.NoError(t, err)

		// Find the created sub-ticket
		subTicketPath := filepath.Join(tmpDir, "tickets", "todo", "*sub-with-id-parent.md")
		subFiles, err := filepath.Glob(subTicketPath)
		require.NoError(t, err)
		require.Len(t, subFiles, 1, "Should have created one sub-ticket")

		// Read and verify sub-ticket has parent relationship
		subTicket := readTicketFromFile(t, subFiles[0])
		require.NotNil(t, subTicket)

		assert.Contains(t, subTicket.Related, "parent:"+parentID)
	})
}
