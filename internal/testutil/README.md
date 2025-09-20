# Test Utilities Package

This package provides common test utilities and helpers to reduce code duplication across test files in the ticketflow project.

## Overview

The testutil package is organized into several categories of helpers:

- **Fixtures** (`fixtures.go`) - Factory functions for creating test data
- **Git Helpers** (`git.go`) - Git repository setup and operations, including git configuration
- **Filesystem** (`filesystem.go`) - File and directory operations
- **Constants** (`constants.go`) - Shared ticket IDs, timestamps, and other canonical test data
- **Context** (`context.go`) - Context testing utilities
- **Assertions** (`assertions.go`) - Common assertion helpers
- **Output** (`output.go`) - Output capture utilities

## Usage Examples

### Creating Test Tickets

```go
// Create a basic test ticket
ticket := testutil.TicketFixture()

// Create a ticket with custom properties
ticket := testutil.TicketFixture(
    testutil.WithID("custom-id"),
    testutil.WithDescription("Custom description"),
    testutil.WithStatus(ticket.StatusDoing),
    testutil.WithPriority(2),
)

// Generate a unique ticket ID
ticketID := testutil.GenerateTicketID(t, "feature")
```

### Setting Up Git Repositories

```go
// Create a test git repository
tmpDir := testutil.CreateTempDir(t)
repo := testutil.SetupGitRepo(t, tmpDir)

// Add a commit
repo.AddCommit(t, "test.txt", "content", "Add test file")

// Create and checkout a branch
repo.CreateBranch(t, "feature-branch")

// IMPORTANT:
// - `SetupGitRepo` always configures git locally (never global)
// - The default branch is forced to "main"
// - Commit signing is disabled to keep tests non-interactive
```

### Bootstrapping Ticketflow Projects

```go
tmpDir := testutil.CreateTempDir(t)

// Create config, tickets directories, tickets/.current, and an initialized git repo
repo := testutil.SetupTicketflowRepo(t, tmpDir)

// You can immediately commit or branch without extra git wiring
repo.AddCommit(t, ".", "", "Bootstrap ticketflow project")
```

### Filesystem Operations

```go
// Create a complete ticketflow project (returns *GitRepo when git init enabled)
tmpDir := testutil.CreateTempDir(t)
repo := testutil.SetupTicketflowProject(t, tmpDir)

// Create a ticket file
ticketPath := testutil.CreateTicketFile(t, tmpDir, "test-123", "todo")

// Assert file operations
testutil.AssertFileExists(t, ticketPath)
testutil.AssertDirExists(t, filepath.Join(tmpDir, "tickets"))

// Change directory with automatic cleanup
testutil.ChDir(t, tmpDir)
```

### Context Testing

```go
// Test with cancelled context
ctx := testutil.CancelledContext()
err := someOperation(ctx)
testutil.AssertContextError(t, err)

// Test with timeout
ctx, cancel := testutil.ShortTimeoutContext()
defer cancel()
err := longRunningOperation(ctx)
testutil.AssertTimeoutError(t, err)
```

### Output Testing

```go
// Capture output
output := testutil.NewOutputCapture()
writer := cli.NewOutputWriter(output.StdoutWriter(), output.StderrWriter(), cli.FormatText)

// Run code that produces output...

// Assert output
testutil.AssertOutputContains(t, output.Stdout(), "expected text")
testutil.AssertOutputEmpty(t, output.Stderr())

// Test JSON output
result := testutil.AssertJSONOutput(t, output.Stdout(), "id", "status")
```

## Best Practices

1. **Git Configuration**: Always use local git configuration in tests, never global
   - Use `testutil.ConfigureGitLocally()` or `testutil.SetupGitRepo()`
   - Never use `git config --global` in tests

2. **Cleanup**: Most helpers automatically register cleanup functions
   - `CreateTempDir` removes the directory on test completion
   - `ChDir` restores the original directory

3. **Parallel Tests**: Be careful with tests that use `os.Chdir()`
   - These tests cannot run in parallel
   - Integration tests typically cannot use `t.Parallel()`

4. **Test Data**: Use fixture functions for consistency
   ```go
   // Good
   ticket := testutil.TicketFixture(testutil.WithStatus(ticket.StatusDoing))
   
   // Avoid
   ticket := &ticket.Ticket{...} // Manual construction
   ```

## Migration Guide

When refactoring existing tests to use these utilities:

1. Replace manual git/config scaffolding with `testutil.SetupTicketflowRepo()` (or combine `SetupTicketflowProject` and `SetupGitRepoWithOptions()` when you need custom behaviour).
2. Replace ticket creation with `testutil.TicketFixture()`
3. Prefer canonical fixtures/constants from `constants.go` (`testutil.TestTicketID`, `testutil.TestCreatedTime`, etc.) instead of redefining literals.
4. Replace file assertions with `testutil.AssertFileExists()` etc.
5. Replace output capture with `testutil.NewOutputCapture()`

## Adding New Helpers

When adding new test helpers:

1. Group related functions in the appropriate file
2. Use descriptive names that clearly indicate the function's purpose
3. Add `t.Helper()` at the beginning of test helper functions
4. Document any important behavior or gotchas
5. Consider if the helper would be useful across multiple test files
