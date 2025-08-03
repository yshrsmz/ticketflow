# Race Condition Fixes for Go Test Suite

## Problem Summary

The ticketflow test suite has two main race conditions:

1. **os.Stdout/os.Stderr capture race**: Multiple parallel tests compete to modify global `os.Stdout` and `os.Stderr` variables
2. **Global state mutation**: Tests call `SetGlobalOutputFormat` which modifies a global variable, causing interference between parallel tests

## Solution: Dependency Injection with OutputWriter

The idiomatic Go solution is to use dependency injection to pass writers instead of relying on global state.

### 1. OutputWriter Pattern

We've introduced an `OutputWriter` struct that encapsulates:
- Output streams (stdout, stderr)
- Output format (text/json)
- Formatting logic

```go
type OutputWriter struct {
    stdout io.Writer
    stderr io.Writer
    format OutputFormat
}
```

### 2. Key Benefits

- **Thread-safe**: Each test gets its own OutputWriter instance
- **No global state**: Format is encapsulated within the writer
- **Testable**: Easy to capture output using `bytes.Buffer`
- **Flexible**: Can redirect output to any `io.Writer`

### 3. Implementation Changes

#### a. Added OutputWriter to cli/output.go

```go
// OutputWriter methods for formatted output
func (w *OutputWriter) PrintJSON(data interface{}) error
func (w *OutputWriter) Printf(format string, args ...interface{})
func (w *OutputWriter) Println(args ...interface{})
func (w *OutputWriter) Error(err error)
```

#### b. Added OutputWriter to App struct

```go
type App struct {
    // ... existing fields ...
    Output *OutputWriter // Output writer for formatted output
}
```

#### c. Added WithOutputWriter option

```go
func WithOutputWriter(writer *OutputWriter) AppOption {
    return func(a *App) {
        a.Output = writer
    }
}
```

### 4. Testing Pattern

Tests can now run in parallel without race conditions:

```go
func TestExample(t *testing.T) {
    t.Parallel() // Safe to run in parallel!
    
    // Create test-specific buffers
    var stdout, stderr bytes.Buffer
    
    // Create test-specific output writer
    outputWriter := cli.NewOutputWriter(&stdout, &stderr, cli.FormatJSON)
    
    // Create app with test output writer
    app, err := cli.NewAppWithOptions(ctx,
        cli.WithWorkingDirectory(tmpDir),
        cli.WithOutputWriter(outputWriter),
    )
    
    // Run your test...
    
    // Check captured output
    assert.Contains(t, stdout.String(), "expected output")
}
```

### 5. Migration Guide

#### For Application Code

Replace direct output calls:

```go
// Before:
fmt.Println("Output")
outputJSON(data)

// After:
app.Output.Println("Output")
app.Output.PrintJSON(data)
```

#### For Error Handling

Replace global error handling:

```go
// Before:
cli.SetGlobalOutputFormat(format)
cli.HandleError(err)

// After:
app.Output.Error(err)
```

#### For Tests

Replace os.Stdout/os.Stderr capture:

```go
// Before (RACE CONDITION):
oldStdout := os.Stdout
r, w, _ := os.Pipe()
os.Stdout = w
// ... run code ...
os.Stdout = oldStdout

// After (THREAD-SAFE):
var stdout bytes.Buffer
outputWriter := cli.NewOutputWriter(&stdout, nil, format)
app, _ := cli.NewAppWithOptions(ctx,
    cli.WithOutputWriter(outputWriter),
)
// ... run code ...
// Check stdout.String()
```

### 6. Backward Compatibility

The existing global functions are preserved for backward compatibility:
- `HandleError()` - still works with global state
- `outputJSON()` - still outputs to os.Stdout
- `SetGlobalOutputFormat()` - still available but should be phased out

### 7. Best Practices

1. **Always use OutputWriter in new code**
2. **Pass OutputWriter through context or as parameter**
3. **Create test-specific writers in tests**
4. **Never modify os.Stdout/os.Stderr in tests**
5. **Use t.Parallel() freely - it's now safe!**

### 8. Additional Patterns to Consider

#### Context-based Output

For deeper call chains, consider passing OutputWriter via context:

```go
type outputKey struct{}

func WithOutput(ctx context.Context, w *OutputWriter) context.Context {
    return context.WithValue(ctx, outputKey{}, w)
}

func GetOutput(ctx context.Context) *OutputWriter {
    if w, ok := ctx.Value(outputKey{}).(*OutputWriter); ok {
        return w
    }
    return NewOutputWriter(nil, nil, FormatText) // default
}
```

#### Interface-based Design

Consider defining an interface for even more flexibility:

```go
type Output interface {
    Printf(format string, args ...interface{})
    Println(args ...interface{})
    PrintJSON(data interface{}) error
    Error(err error)
}
```

This allows for easy mocking and alternative implementations.

### 9. Performance Considerations

- No performance penalty vs direct os.Stdout writes
- Slightly more memory allocation (one OutputWriter per App)
- Enables true parallel test execution (faster overall test suite)

### 10. Next Steps

1. Update all command implementations to use `app.Output`
2. Refactor tests to use OutputWriter pattern
3. Remove global state usage in production code
4. Consider deprecating global output functions
5. Add linting rules to prevent direct os.Stdout usage in new code