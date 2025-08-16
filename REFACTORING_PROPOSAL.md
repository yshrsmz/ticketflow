# Refactoring Proposal: Clean CLI Output Architecture

## Overview
Based on analysis of industry-leading CLI tools (kubectl, docker, Kong), here are three architectural improvements to address the current issues.

## Issue 1: Switch Statement in OutputWriter

### Current Problem
```go
func (w *textResultWriter) PrintResult(data interface{}) error {
    switch v := data.(type) {
    case *CleanupResult:
        // format cleanup result
    case []ticket.Ticket:
        // format ticket list
    // ... grows huge over time
    }
}
```

### Proposed Solution: Printable Interface Pattern (kubectl-style)

```go
// internal/cli/printable.go
type Printable interface {
    TextRepresentation() string
    StructuredData() interface{}
}

// Each result type implements Printable
func (r *CleanupResult) TextRepresentation() string {
    // Returns formatted text
}

func (r *CleanupResult) StructuredData() interface{} {
    // Returns data for JSON serialization
}

// OutputWriter becomes simpler
func (w *textResultWriter) PrintResult(data interface{}) error {
    if p, ok := data.(Printable); ok {
        fmt.Fprint(w.stdout, p.TextRepresentation())
        return nil
    }
    // Fallback for non-Printable types
    return w.printDefault(data)
}
```

**Benefits:**
- Each result type owns its formatting logic (Single Responsibility)
- No central switch statement to maintain
- Easy to add new result types
- Follows kubectl's ResourcePrinter pattern

## Issue 2: App Initialization Order

### Current Problem
```go
app, err := cli.NewApp(ctx)
// ... parse flags ...
// Then mutate app after creation
app.Output = cli.NewOutputWriter(os.Stdout, os.Stderr, outputFormat)
app.StatusWriter = cli.NewStatusWriter(os.Stdout, outputFormat)
```

### Proposed Solution: Builder Pattern with Options (Kong/Docker-style)

```go
// internal/cli/app_builder.go
func NewAppWithFormat(ctx context.Context, format OutputFormat) (*App, error) {
    return NewAppWithOptions(ctx, 
        WithOutputFormat(format),
        WithStatusWriter(NewStatusWriter(os.Stdout, format)),
        WithResultWriter(NewResultWriter(os.Stdout, os.Stderr, format)),
    )
}

// Commands become cleaner
func (c *CleanupCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
    f := flags.(*cleanupFlags)
    
    // App is created with the right format from the start
    app, err := cli.NewAppWithFormat(ctx, cli.ParseOutputFormat(f.format))
    if err != nil {
        return err
    }
    
    // No mutation needed - app is properly initialized
    return app.AutoCleanup(ctx, f.dryRun)
}
```

**Alternative: Command Context Pattern (kubectl-style)**
```go
// Create a context that carries initialized resources
type CommandContext struct {
    *cli.App
    OutputFormat OutputFormat
}

func NewCommandContext(ctx context.Context, format string) (*CommandContext, error) {
    outputFormat := ParseOutputFormat(format)
    app, err := NewAppWithOptions(ctx,
        WithOutputFormat(outputFormat),
    )
    if err != nil {
        return nil, err
    }
    
    return &CommandContext{
        App:          app,
        OutputFormat: outputFormat,
    }, nil
}
```

**Benefits:**
- App is immutable after creation
- Clear initialization flow
- Follows Kong's struct-based approach
- No awkward state mutations

## Issue 3: Test-Specific Nil Checks

### Current Problem
```go
// Production code has test-specific checks
if app.StatusWriter == nil {
    app.StatusWriter = NewNullStatusWriter()
}
```

### Proposed Solution: Always Initialize with Null Object Pattern

```go
// internal/cli/commands.go
func NewAppWithOptions(ctx context.Context, opts ...AppOption) (*App, error) {
    app := &App{
        // Always initialize with defaults (never nil)
        StatusWriter: NewNullStatusWriter(),
        Output:       NewOutputWriter(nil, nil, FormatText),
    }
    
    // Options can override defaults
    for _, opt := range opts {
        if err := opt(app); err != nil {
            return nil, err
        }
    }
    
    // StatusWriter is never nil
    return app, nil
}

// Tests explicitly set what they need
func TestSomething(t *testing.T) {
    app, _ := NewAppWithOptions(ctx,
        WithStatusWriter(NewTextStatusWriter(buffer)),
    )
    // No nil checks needed
}
```

**Benefits:**
- No test-specific code in production
- Follows Null Object Pattern (docker, git)
- Tests are explicit about their needs
- Cleaner production code

## Migration Path

### Phase 1: Implement Printable Interface (Low Risk)
1. Add Printable interface
2. Implement for existing result types
3. Update OutputWriter to check for Printable first
4. Keep switch as fallback for backward compatibility

### Phase 2: Fix Initialization Order (Medium Risk)
1. Add NewAppWithFormat helper function
2. Update commands one by one to use it
3. Remove post-creation mutations

### Phase 3: Remove Nil Checks (Low Risk)  
1. Ensure NewAppWithOptions always initializes StatusWriter
2. Update tests to always provide StatusWriter
3. Remove nil checks from production code

## Expected Outcomes

1. **Cleaner Architecture**: Following patterns from kubectl, docker, Kong
2. **Better Testability**: No special production paths for tests
3. **Maintainability**: Each result type owns its formatting
4. **Scalability**: Easy to add new commands and output types
5. **Go Idiomatic**: Small, focused interfaces; composition over inheritance

## References
- kubectl: ResourcePrinter interface pattern
- docker: Structured result types with formatters
- Kong: Struct-based CLI with clean initialization
- git: Plumbing vs porcelain command separation