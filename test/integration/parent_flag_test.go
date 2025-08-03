package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	
	// Create test directory
	tmpDir := t.TempDir()
	
	// Save current directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err)
	}()
	
	// Change to test directory
	require.NoError(t, os.Chdir(tmpDir))
	
	// Initialize git repo
	runCommand(t, exec.Command("git", "init"))
	runCommand(t, exec.Command("git", "config", "user.name", "Test User"))
	runCommand(t, exec.Command("git", "config", "user.email", "test@example.com"))
	
	// Create a README to have something to commit
	err = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test Project"), 0644)
	require.NoError(t, err)
	
	runCommand(t, exec.Command("git", "add", "-A"))
	runCommand(t, exec.Command("git", "commit", "-m", "Initial commit"))
	
	// Initialize ticketflow
	_, err = exec.Command("ticketflow", "init").CombinedOutput()
	require.NoError(t, err)
	
	// Create a parent ticket first
	parentOut, err := exec.Command("ticketflow", "new", "parent-feature").CombinedOutput()
	require.NoError(t, err, "Failed to create parent ticket: %s", parentOut)
	
	// Get the parent ticket ID
	parentTicketPath := filepath.Join(tmpDir, "tickets", "todo", "*parent-feature.md")
	parentFiles, err := filepath.Glob(parentTicketPath)
	require.NoError(t, err)
	require.Len(t, parentFiles, 1, "Should have created one parent ticket")
	parentID := strings.TrimSuffix(filepath.Base(parentFiles[0]), ".md")
	
	t.Run("create sub-ticket with --parent flag", func(t *testing.T) {
		// Create sub-ticket with explicit parent
		out, err := exec.Command("ticketflow", "new", "sub-feature", "--parent", parentID).CombinedOutput()
		require.NoError(t, err, "Failed to create sub-ticket: %s", out)
		
		// Verify output mentions parent
		assert.Contains(t, string(out), "Creating sub-ticket with parent: "+parentID)
		
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
		// Create sub-ticket with short form parent flag
		out, err := exec.Command("ticketflow", "new", "another-sub", "-p", parentID).CombinedOutput()
		require.NoError(t, err, "Failed to create sub-ticket: %s", out)
		
		// Verify output mentions parent
		assert.Contains(t, string(out), "Creating sub-ticket with parent: "+parentID)
		
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
		out, err := exec.Command("ticketflow", "new", "orphan-sub", "--parent", "non-existent-ticket").CombinedOutput()
		assert.Error(t, err, "Should fail with non-existent parent")
		assert.Contains(t, string(out), "Parent ticket not found")
	})
	
	t.Run("error on self-parent", func(t *testing.T) {
		// Try to create ticket with itself as parent
		out, err := exec.Command("ticketflow", "new", "self-parent", "--parent", "self-parent").CombinedOutput()
		assert.Error(t, err, "Should fail with self-parent")
		assert.Contains(t, string(out), "cannot be its own parent")
	})
	
	t.Run("explicit parent overrides implicit worktree parent", func(t *testing.T) {
		// Start working on parent ticket to create worktree
		out, err := exec.Command("ticketflow", "start", parentID).CombinedOutput()
		require.NoError(t, err, "Failed to start parent ticket: %s", out)
		
		// Change to parent worktree
		parentWorktreePath := filepath.Join(tmpDir, "..", "ticketflow.worktrees", parentID)
		require.NoError(t, os.Chdir(parentWorktreePath))
		
		// Create another parent ticket in main repo
		require.NoError(t, os.Chdir(tmpDir))
		out, err = exec.Command("ticketflow", "new", "another-parent").CombinedOutput()
		require.NoError(t, err, "Failed to create another parent: %s", out)
		
		// Get the another parent ticket ID
		anotherParentPath := filepath.Join(tmpDir, "tickets", "todo", "*another-parent.md")
		anotherParentFiles, err := filepath.Glob(anotherParentPath)
		require.NoError(t, err)
		require.Len(t, anotherParentFiles, 1)
		anotherParentID := strings.TrimSuffix(filepath.Base(anotherParentFiles[0]), ".md")
		
		// Go back to first parent's worktree
		require.NoError(t, os.Chdir(parentWorktreePath))
		
		// Create sub-ticket with explicit parent (different from current worktree)
		out, err = exec.Command("ticketflow", "new", "explicit-over-implicit", "--parent", anotherParentID).CombinedOutput()
		require.NoError(t, err, "Failed to create sub-ticket: %s", out)
		
		// Verify warning about using explicit parent
		assert.Contains(t, string(out), "Using explicit parent '"+anotherParentID+"' instead of current worktree '"+parentID+"'")
		
		// Find the created sub-ticket
		subTicketPath := filepath.Join(parentWorktreePath, "tickets", "todo", "*explicit-over-implicit.md")
		subFiles, err := filepath.Glob(subTicketPath)
		require.NoError(t, err)
		require.Len(t, subFiles, 1)
		
		// Read and verify sub-ticket has explicit parent, not implicit
		subTicket := readTicketFromFile(t, subFiles[0])
		
		require.NotNil(t, subTicket.Related)
		assert.Contains(t, subTicket.Related, "parent:"+anotherParentID)
		assert.NotContains(t, subTicket.Related, "parent:"+parentID)
	})
}

func TestNewCommandWithBothParentFlags(t *testing.T) {
	// Integration tests cannot run in parallel due to os.Chdir
	
	// Create test directory
	tmpDir := t.TempDir()
	
	// Save current directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err)
	}()
	
	// Change to test directory
	require.NoError(t, os.Chdir(tmpDir))
	
	// Initialize git repo
	runCommand(t, exec.Command("git", "init"))
	runCommand(t, exec.Command("git", "config", "user.name", "Test User"))
	runCommand(t, exec.Command("git", "config", "user.email", "test@example.com"))
	
	// Create a README to have something to commit
	err = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test Project"), 0644)
	require.NoError(t, err)
	
	runCommand(t, exec.Command("git", "add", "-A"))
	runCommand(t, exec.Command("git", "commit", "-m", "Initial commit"))
	
	// Initialize ticketflow
	_, err = exec.Command("ticketflow", "init").CombinedOutput()
	require.NoError(t, err)
	
	// Create parent tickets
	out, err := exec.Command("ticketflow", "new", "parent1").CombinedOutput()
	require.NoError(t, err, "Failed to create parent1: %s", out)
	
	out, err = exec.Command("ticketflow", "new", "parent2").CombinedOutput()
	require.NoError(t, err, "Failed to create parent2: %s", out)
	
	// Get parent IDs
	parent1Path := filepath.Join(tmpDir, "tickets", "todo", "*parent1.md")
	parent1Files, err := filepath.Glob(parent1Path)
	require.NoError(t, err)
	require.Len(t, parent1Files, 1)
	parent1ID := strings.TrimSuffix(filepath.Base(parent1Files[0]), ".md")
	
	parent2Path := filepath.Join(tmpDir, "tickets", "todo", "*parent2.md")
	parent2Files, err := filepath.Glob(parent2Path)
	require.NoError(t, err)
	require.Len(t, parent2Files, 1)
	parent2ID := strings.TrimSuffix(filepath.Base(parent2Files[0]), ".md")
	
	t.Run("error when both --parent and -p are used with different values", func(t *testing.T) {
		// Try to create ticket with both parent flags
		out, err := exec.Command("ticketflow", "new", "conflicting-parents", "--parent", parent1ID, "-p", parent2ID).CombinedOutput()
		assert.Error(t, err, "Should fail when both parent flags have different values")
		assert.Contains(t, string(out), "cannot specify both --parent and -p flags")
	})
	
	t.Run("success when both --parent and -p have same value", func(t *testing.T) {
		// Create ticket with both parent flags having same value
		out, err := exec.Command("ticketflow", "new", "same-parents", "--parent", parent1ID, "-p", parent1ID).CombinedOutput()
		require.NoError(t, err, "Should succeed when both flags have same value: %s", out)
		
		// Verify ticket was created with correct parent
		ticketPath := filepath.Join(tmpDir, "tickets", "todo", "*same-parents.md")
		ticketFiles, err := filepath.Glob(ticketPath)
		require.NoError(t, err)
		require.Len(t, ticketFiles, 1)
		
		createdTicket := readTicketFromFile(t, ticketFiles[0])
		assert.Contains(t, createdTicket.Related, "parent:"+parent1ID)
	})
}

// runCommand is a helper to run a command and fail the test if it errors
func runCommand(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "Command failed: %s\nOutput: %s", cmd.String(), out)
}