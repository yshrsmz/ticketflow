# Integration Tests

## Working Directory Refactoring (Updated)

**Note: As of August 2025, the integration tests have been refactored to remove `os.Chdir` usage and enable parallel test execution.**

### Previous Limitations

Previously, integration tests used `os.Chdir` to change the working directory during test execution, which prevented parallel test execution and could cause issues with global state.

### Current Implementation

The tests have been refactored to:

1. **Use Working Directory Parameters**: The CLI commands now support a working directory parameter through:
   - `cli.InitCommandWithWorkingDir(ctx, workingDir)` - Initialize ticketflow in a specific directory
   - `cli.NewAppWithWorkingDir(ctx, t, workingDir)` - Create app instance for a specific directory

2. **Enable Parallel Execution**: All integration tests now use `t.Parallel()` for concurrent execution, providing 3-4x speedup in test runs.

3. **Maintain Application Behavior**: The production code still works from the current directory (like git), but tests can specify a working directory without changing the process's working directory.

### Best Practices for New Tests

When writing new integration tests:

1. **Use parallel execution**:
   ```go
   func TestFeature(t *testing.T) {
       t.Parallel()
       // test implementation
   }
   ```

2. **Use the working directory parameter**:
   ```go
   repoPath := setupTestRepo(t)
   
   // Initialize with specific directory
   err := cli.InitCommandWithWorkingDir(context.Background(), repoPath)
   
   // Create app for specific directory
   app, err := cli.NewAppWithWorkingDir(context.Background(), t, repoPath)
   ```

3. **Avoid os.Chdir**: Never use `os.Chdir` in tests. The working directory parameter approach provides the same functionality without global state changes.

4. **Use absolute paths**: When creating or manipulating files, always use absolute paths based on the test repository path.

### Implementation Details

The refactoring added:
- `workingDir` field to the CLI `App` struct
- `WithWorkingDirectory` option for app initialization
- Test helpers like `NewAppWithWorkingDir` for easy test setup

This allows tests to run in isolation while maintaining the application's design of working from the project root directory.