# Testing Patterns for TicketFlow

This document describes the testing patterns and best practices for writing tests in the TicketFlow project.

## Overview

TicketFlow tests have been refactored to support parallel execution by removing `os.Chdir` usage. This provides significant performance improvements (3-4x speedup) while maintaining test isolation.

## Key Principles

1. **No os.Chdir**: Never use `os.Chdir` in tests as it modifies global state
2. **Parallel Execution**: Use `t.Parallel()` to enable concurrent test execution
3. **Working Directory Parameters**: Use the working directory parameter approach
4. **Absolute Paths**: Always use absolute paths for file operations

## Unit Tests

### Basic Pattern

```go
func TestFeature(t *testing.T) {
    t.Parallel()
    
    // Test implementation
}
```

### Testing CLI Commands

```go
func TestCLICommand(t *testing.T) {
    t.Parallel()
    
    // Create temporary directory
    tmpDir := t.TempDir()
    
    // Initialize with working directory
    err := cli.InitCommandWithWorkingDir(context.Background(), tmpDir)
    require.NoError(t, err)
    
    // Create app with working directory
    app, err := cli.NewAppWithWorkingDir(context.Background(), t, tmpDir)
    require.NoError(t, err)
    
    // Test the command
    err = app.NewTicket(context.Background(), "test-feature", cli.FormatText)
    assert.NoError(t, err)
}
```

## Integration Tests

### Setting Up Test Repository

```go
func TestIntegration(t *testing.T) {
    t.Parallel()
    
    // Setup test repository
    repoPath := setupTestRepo(t)
    
    // Initialize ticketflow
    err := cli.InitCommandWithWorkingDir(context.Background(), repoPath)
    require.NoError(t, err)
    
    // Create app instance
    app, err := cli.NewAppWithWorkingDir(context.Background(), t, repoPath)
    require.NoError(t, err)
    
    // Run integration test
}
```

### Working with Git

```go
// Create git client for specific path
gitCmd := git.New(repoPath)

// Execute git commands
_, err := gitCmd.Exec(context.Background(), "add", ".")
require.NoError(t, err)

_, err = gitCmd.Exec(context.Background(), "commit", "-m", "Test commit")
require.NoError(t, err)
```

### File Operations

Always use absolute paths:

```go
// Good - absolute path
configPath := filepath.Join(repoPath, ".ticketflow.yaml")
err := os.WriteFile(configPath, data, 0644)

// Bad - relative path
err := os.WriteFile(".ticketflow.yaml", data, 0644)
```

## Test Helpers

### NewAppWithWorkingDir

Use this helper to create an app instance for a specific directory:

```go
app, err := cli.NewAppWithWorkingDir(context.Background(), t, workingDir)
```

### InitCommandWithWorkingDir

Use this to initialize ticketflow in a specific directory:

```go
err := cli.InitCommandWithWorkingDir(context.Background(), workingDir)
```

## Common Patterns

### Testing with Worktrees

```go
// Enable worktrees in config
cfg, err := config.Load(repoPath)
require.NoError(t, err)
cfg.Worktree.Enabled = true
cfg.Worktree.BaseDir = "./.worktrees"
err = cfg.Save(filepath.Join(repoPath, ".ticketflow.yaml"))
require.NoError(t, err)

// Create app with updated config
app, err := cli.NewAppWithWorkingDir(context.Background(), t, repoPath)
require.NoError(t, err)
```

### Verifying File Creation

```go
// Check file exists
ticketPath := filepath.Join(repoPath, "tickets", "todo", ticketID+".md")
_, err := os.Stat(ticketPath)
assert.NoError(t, err)

// Check file doesn't exist
_, err = os.Stat(oldPath)
assert.True(t, os.IsNotExist(err))
```

### Git Configuration in Tests

Always configure git locally in test directories:

```go
// Configure git for test repo
gitCmd := git.New(repoPath)
_, err = gitCmd.Exec(context.Background(), "config", "user.name", "Test User")
require.NoError(t, err)
_, err = gitCmd.Exec(context.Background(), "config", "user.email", "test@example.com")
require.NoError(t, err)
```

**Warning**: Never use `git config --global` in tests as it modifies the user's git configuration.

## Migration Guide

If you're updating old tests that use `os.Chdir`:

### Before (old pattern):
```go
func TestOldPattern(t *testing.T) {
    repoPath := setupTestRepo(t)
    originalWd, err := os.Getwd()
    require.NoError(t, err)
    defer func() {
        err := os.Chdir(originalWd)
        require.NoError(t, err)
    }()
    err = os.Chdir(repoPath)
    require.NoError(t, err)
    
    err = cli.InitCommand(context.Background())
    require.NoError(t, err)
    
    app, err := cli.NewApp(context.Background())
    require.NoError(t, err)
}
```

### After (new pattern):
```go
func TestNewPattern(t *testing.T) {
    t.Parallel()
    
    repoPath := setupTestRepo(t)
    
    err := cli.InitCommandWithWorkingDir(context.Background(), repoPath)
    require.NoError(t, err)
    
    app, err := cli.NewAppWithWorkingDir(context.Background(), t, repoPath)
    require.NoError(t, err)
}
```

## Performance Benefits

The refactoring to remove `os.Chdir` enables:
- Parallel test execution with `t.Parallel()`
- 3-4x speedup in test suite execution
- Better test isolation and reliability
- No global state modifications

## Troubleshooting

### Tests Creating Files in Wrong Directory

If tests create files in the source tree instead of temp directories:
1. Ensure all paths are absolute (use `filepath.Join(repoPath, ...)`)
2. Check that the app was created with `NewAppWithWorkingDir`
3. Verify that git operations use `git.New(repoPath)`

### Parallel Test Failures

If tests fail when run in parallel but pass individually:
1. Check for shared state or resources
2. Ensure each test uses its own temp directory
3. Verify no hardcoded paths or ports

## Summary

The key to writing good tests in TicketFlow is:
1. Always use `t.Parallel()` for better performance
2. Use working directory parameters instead of `os.Chdir`
3. Work with absolute paths
4. Configure git locally in test directories
5. Use the provided test helpers

Following these patterns ensures tests are fast, reliable, and maintainable.