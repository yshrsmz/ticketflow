# Command Implementation Migration Guide

This guide shows how to update command implementations to use the OutputWriter pattern.

## Example: Migrating the NewTicket Command

### Before (using global os.Stdout)

```go
func (app *App) NewTicket(ctx context.Context, slug string, format OutputFormat) error {
    // Create ticket...
    
    if format == FormatJSON {
        return outputJSON(map[string]interface{}{
            "ticket": ticketToJSON(t, ""),
            "message": fmt.Sprintf("Created ticket: %s", t.ID),
        })
    }
    
    fmt.Printf("Created ticket: %s\n", t.ID)
    fmt.Printf("Path: %s\n", t.Path)
    return nil
}
```

### After (using OutputWriter)

```go
func (app *App) NewTicket(ctx context.Context, slug string) error {
    // Create ticket...
    
    if app.Output.GetFormat() == FormatJSON {
        return app.Output.PrintJSON(map[string]interface{}{
            "ticket": ticketToJSON(t, ""),
            "message": fmt.Sprintf("Created ticket: %s", t.ID),
        })
    }
    
    app.Output.Printf("Created ticket: %s\n", t.ID)
    app.Output.Printf("Path: %s\n", t.Path)
    return nil
}
```

## Example: Migrating the ListTickets Command

### Before

```go
func (app *App) ListTickets(ctx context.Context, status Status, limit int, format OutputFormat) error {
    tickets, err := app.Manager.List(ctx, status, limit)
    if err != nil {
        return err
    }
    
    if format == FormatJSON {
        return outputJSON(map[string]interface{}{
            "tickets": tickets,
        })
    }
    
    // Text output
    for _, t := range tickets {
        fmt.Printf("%-20s %s\n", t.ID, t.Description)
    }
    return nil
}
```

### After

```go
func (app *App) ListTickets(ctx context.Context, status Status, limit int) error {
    tickets, err := app.Manager.List(ctx, status, limit)
    if err != nil {
        return err
    }
    
    if app.Output.GetFormat() == FormatJSON {
        return app.Output.PrintJSON(map[string]interface{}{
            "tickets": tickets,
        })
    }
    
    // Text output
    for _, t := range tickets {
        app.Output.Printf("%-20s %s\n", t.ID, t.Description)
    }
    return nil
}
```

## Key Changes

1. **Remove format parameter**: The format is now stored in `app.Output`
2. **Replace fmt.Printf**: Use `app.Output.Printf`
3. **Replace outputJSON**: Use `app.Output.PrintJSON`
4. **Error handling**: Errors can be handled via `app.Output.Error(err)`

## Command Handler Updates

In the main.go command handlers:

### Before

```go
func handleNew(ctx context.Context, args []string, outputFormat string) error {
    // Parse args...
    
    app, err := cli.NewApp(ctx)
    if err != nil {
        cli.HandleError(err)
        return err
    }
    
    format := cli.ParseOutputFormat(outputFormat)
    err = app.NewTicket(ctx, slug, format)
    if err != nil {
        cli.HandleError(err)
        return err
    }
    return nil
}
```

### After

```go
func handleNew(ctx context.Context, args []string, outputFormat string) error {
    // Parse args...
    
    format := cli.ParseOutputFormat(outputFormat)
    outputWriter := cli.NewOutputWriter(nil, nil, format)
    
    app, err := cli.NewAppWithOptions(ctx,
        cli.WithOutputWriter(outputWriter),
    )
    if err != nil {
        outputWriter.Error(err)
        return err
    }
    
    err = app.NewTicket(ctx, slug)
    if err != nil {
        outputWriter.Error(err)
        return err
    }
    return nil
}
```

## Testing Command Implementations

When testing commands that use OutputWriter:

```go
func TestCommand(t *testing.T) {
    // Create test buffers
    var stdout, stderr bytes.Buffer
    
    // Create test output writer
    outputWriter := cli.NewOutputWriter(&stdout, &stderr, cli.FormatJSON)
    
    // Create app with test writer
    app, _ := cli.NewAppWithOptions(ctx,
        cli.WithOutputWriter(outputWriter),
    )
    
    // Run command
    err := app.SomeCommand(ctx)
    
    // Check output
    assert.Contains(t, stdout.String(), "expected output")
    
    // Check errors
    if err != nil {
        assert.Contains(t, stderr.String(), "error message")
    }
}
```

## Benefits

1. **No race conditions**: Each command instance has its own output writer
2. **Testable**: Easy to capture and verify output
3. **Flexible**: Can redirect output anywhere (files, network, etc.)
4. **Consistent**: All output goes through the same interface
5. **Format encapsulation**: Format is part of the writer, not passed around