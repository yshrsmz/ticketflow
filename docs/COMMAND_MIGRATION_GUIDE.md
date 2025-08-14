# Command Interface Migration Guide

> **Status**: Version & Help commands migrated ✅ | Other commands pending migration

This guide documents the migration from the switch-based command system to the new Command interface.

## Current Architecture

The current system uses:
1. A large switch statement in `main.go` (lines 143-328)
2. A `Command` struct in `cmd/ticketflow/command.go` 
3. The `parseAndExecute` function to handle common patterns
4. Handler functions like `handleNew`, `handleList`, etc. in `internal/cli/commands.go`

## Migration Strategy

### Phase 1: Parallel Systems (Recommended)

Run both systems in parallel to allow incremental migration:

```go
// In main.go
func main() {
    // Try new registry first
    if cmd, ok := commandRegistry.Get(os.Args[1]); ok {
        return executeNewCommand(ctx, cmd, os.Args[2:])
    }
    
    // Fall back to old switch for unmigrated commands
    switch os.Args[1] {
    case "list":
        // old implementation
    // ... other unmigrated commands
    }
}
```

### Phase 2: Command Migration Order

Migrate commands from simplest to most complex:

1. **Start with stateless commands** (no dependencies):
   - `version` - No flags, no app dependency
   - `help` - Simple output

2. **Then simple commands** (minimal flags):
   - `init` - No flags, simple execution
   - `status` - Simple flags, read-only

3. **Then standard CRUD commands**:
   - `list` - Flags but straightforward
   - `show` - Single argument
   - `new` - More complex with parent handling

4. **Finally complex commands**:
   - `start` - Worktree management
   - `close` - Multiple execution paths
   - `cleanup` - Interactive prompts
   - `worktree` - Subcommands

## Step-by-Step Migration Example

### Step 1: Create Command Implementation

```go
// internal/cli/commands/version.go
package commands

import (
    "context"
    "flag"
    "fmt"
    "github.com/yshrsmz/ticketflow/internal/command"
)

type VersionCommand struct {
    Version   string
    GitCommit string
    BuildTime string
}

func NewVersionCommand(version, commit, buildTime string) command.Command {
    return &VersionCommand{
        Version:   version,
        GitCommit: commit,
        BuildTime: buildTime,
    }
}

func (c *VersionCommand) Name() string        { return "version" }
func (c *VersionCommand) Description() string { return "Show version information" }
func (c *VersionCommand) Usage() string       { return "version" }

func (c *VersionCommand) SetupFlags(fs *flag.FlagSet) interface{} {
    return nil // No flags
}

func (c *VersionCommand) Validate(flags interface{}, args []string) error {
    return nil // No validation needed
}

func (c *VersionCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
    fmt.Printf("ticketflow version %s\n", c.Version)
    if c.Version != "dev" || c.GitCommit != "unknown" {
        fmt.Printf("  Git commit: %s\n", c.GitCommit)
        fmt.Printf("  Built at:   %s\n", c.BuildTime)
    }
    return nil
}
```

### Step 2: Register Command

```go
// cmd/ticketflow/main.go
var commandRegistry = command.NewRegistry()

func init() {
    // Register commands
    commandRegistry.Register(commands.NewVersionCommand(Version, GitCommit, BuildTime))
    // Add more as they're migrated...
}
```

### Step 3: Update Main Dispatcher

```go
// cmd/ticketflow/main.go
func runCommand(ctx context.Context) error {
    // ... existing validation ...
    
    // Try new registry first
    if cmd, ok := commandRegistry.Get(os.Args[1]); ok {
        return executeNewCommand(ctx, cmd, os.Args[2:])
    }
    
    // Fall back to old switch
    switch os.Args[1] {
    case "version", "-v", "--version":
        // REMOVED - handled by registry now
    case "list":
        // Still using old system
        return parseAndExecute(ctx, Command{...}, os.Args[2:])
    // ... rest of switch
    }
}

func executeNewCommand(ctx context.Context, cmd command.Command, args []string) error {
    fs := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
    
    // Setup flags
    var flags interface{}
    if cmd.SetupFlags != nil {
        flags = cmd.SetupFlags(fs)
    }
    
    // Add logging flags
    loggingOpts := cli.AddLoggingFlags(fs)
    
    // Parse
    if err := fs.Parse(args); err != nil {
        return err
    }
    
    // Configure logging
    if err := cli.ConfigureLogging(loggingOpts); err != nil {
        return err
    }
    
    // Validate
    if err := cmd.Validate(flags, fs.Args()); err != nil {
        return err
    }
    
    // Execute
    return cmd.Execute(ctx, flags, fs.Args())
}
```

## Handling Special Cases

### 1. Commands with Subcommands (worktree)

```go
type WorktreeCommand struct {
    subcommands map[string]command.Command
}

func (c *WorktreeCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
    if len(args) == 0 {
        c.printUsage()
        return nil
    }
    
    if subcmd, ok := c.subcommands[args[0]]; ok {
        return subcmd.Execute(ctx, flags, args[1:])
    }
    
    return fmt.Errorf("unknown worktree command: %s", args[0])
}
```

### 2. Commands with App Dependencies (Updated Pattern)

```go
// internal/cli/commands/close.go
type CloseCommand struct{}

func NewCloseCommand() command.Command {
    return &CloseCommand{}
}

func (c *CloseCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
    app, err := cli.NewApp(ctx)
    if err != nil {
        return err
    }
    
    f := flags.(*closeFlags)
    
    // App methods now return entities
    ticket, err := app.CloseTicket(ctx, f.force)
    if err != nil {
        return err
    }
    
    // Handle JSON output if requested
    if f.format == FormatJSON {
        output := map[string]interface{}{
            "ticket_id": ticket.ID,
            "status":    string(ticket.Status()),
            "closed_at": ticket.ClosedAt.Time.Format(time.RFC3339),
            "duration":  cli.FormatDuration(cli.CalculateDuration(ticket)),
            "parent":    cli.ExtractParentID(ticket),
        }
        return app.Output.PrintJSON(output)
    }
    
    return nil // Text output handled by App method
}
```

### 3. Testing Migrated Commands

```go
func TestListCommand(t *testing.T) {
    cmd := &ListCommand{
        appFactory: func(ctx context.Context) (*cli.App, error) {
            // Return mock app
            return mockApp, nil
        },
    }
    
    // Test flag setup
    fs := flag.NewFlagSet("test", flag.ContinueOnError)
    flags := cmd.SetupFlags(fs)
    
    // Test validation
    err := cmd.Validate(flags, []string{})
    assert.NoError(t, err)
    
    // Test execution
    err = cmd.Execute(context.Background(), flags, []string{})
    assert.NoError(t, err)
}
```

## Benefits After Migration

1. **Better Organization**: Each command in its own file
2. **Easier Testing**: Commands can be tested in isolation
3. **Better Dependency Injection**: Commands explicitly declare dependencies
4. **Cleaner Main**: No 200+ line switch statement
5. **Plugin Ready**: Registry pattern makes it easy to add external commands
6. **Type Safety**: Compile-time checking of command implementations

## Migration Status

### Completed ✅
- [x] Create `internal/cli/commands/` directory
- [x] Implement registry initialization in main.go
- [x] Add executeNewCommand function
- [x] **version** command (including -v, --version aliases)
- [x] **help** command (including -h, --help aliases)
- [x] **init** - Initialize ticket system
- [x] **status** - Show current ticket status (first command with App dependency)
- [x] **list** - List tickets with filters (including ls alias)
- [x] **show** - Show ticket details (first command with positional arguments)
- [x] **new** - Create new ticket (first state-changing command, supports --parent/-p and --format/-o flags)
- [x] **start** - Start working on ticket (worktree management, supports --force/-f and --format/-o flags, JSON output)
- [x] **close** - Close current/specified ticket (dual-mode operation, supports --force/-f, --reason, and --format/-o flags, JSON output)

### In Progress 🚧
- [ ] Create migration tickets for remaining commands

### Pending Migration 📋

#### Simple Commands (No Dependencies)

#### Read-Only Commands

#### State-Changing Commands
- [ ] **restore** - Restore closed ticket

#### Complex Commands
- [ ] **worktree** - Manage git worktrees
- [ ] **cleanup** - Clean up worktrees and branches
- [ ] **migrate** - Migrate ticket dates

### Final Cleanup
- [ ] Remove old Command struct from command.go
- [ ] Remove parseAndExecute function
- [ ] Remove switch statement from main.go
- [ ] Update all documentation

## Rollback Plan

If issues arise during migration:
1. Commands can be moved back to switch statement immediately
2. The registry can be disabled by commenting out the registry check
3. Both systems can coexist indefinitely if needed

## Common Pitfalls

1. **Don't forget logging flags**: The new executeNewCommand needs to handle logging
2. **Test argument handling**: Make sure fs.Args() is used correctly
3. **Handle aliases**: Commands like "version", "-v", "--version" need mapping
4. **Preserve error messages**: Keep the same user-facing error messages
5. **Watch for nil interfaces**: SetupFlags can return nil for commands without flags

## New Patterns Established

### App Methods Return Primary Entities (✅ Completed 2025-08-14)
**Background**: App methods now return the primary entity they operate on, eliminating the need for commands to re-fetch data. This refactoring was completed in ticket 250814-121422.

**Benefits**:
- **Performance**: 50% fewer file I/O operations (no re-fetching)
- **Better testability**: Can assert on returned values directly
- **Cleaner command code**: Commands focus on presentation, not data retrieval
- **Idiomatic Go**: Follows standard `(T, error)` return pattern

**Method Signatures**:
- `CloseTicket(ctx, force) (*ticket.Ticket, error)`
- `CloseTicketWithReason(ctx, reason, force) (*ticket.Ticket, error)`
- `CloseTicketByID(ctx, ticketID, reason, force) (*ticket.Ticket, error)`
- `StartTicket(ctx, ticketID, force) (*StartTicketResult, error)` ⚠️ See note below
- `NewTicket(ctx, slug, parent) (*ticket.Ticket, error)`
- `RestoreCurrentTicket(ctx) (*ticket.Ticket, error)`

**Special Case: StartTicketResult**
The `StartTicket` method returns a custom struct instead of just the ticket:
```go
type StartTicketResult struct {
    Ticket               *ticket.Ticket  // The started ticket
    WorktreePath         string          // Path to created worktree
    ParentBranch         string          // Branch it was created from
    InitCommandsExecuted bool            // Whether init commands ran
}
```
**Reasoning**: StartTicket orchestrates a complex workflow (worktree creation, branch management, init commands) and all this information is needed by commands for output. Returning just the ticket would require additional git queries, defeating our performance goals.

**Helper Methods** (in `internal/cli/helpers.go`):
- `CalculateDuration(ticket)` - Calculate work duration (handles nil and invalid states)
- `ExtractParentID(ticket)` - Get parent from Related field (nil-safe)
- `FormatDuration(duration)` - Human-readable duration format (e.g., "2h30m")

### Dual-Mode Commands (from close command)
Commands that can operate with or without arguments (0 or 1 args):
- Store args in flags struct during Validate for Execute to use
- Different behavior based on argument presence
- Example: `close` - no args closes current ticket, with ID closes specific ticket

### JSON Output for State-Changing Commands (Updated Pattern)
App methods now return the primary entity they operate on:
1. Execute the operation and receive the entity
2. Use the returned entity directly for JSON output (no re-fetching needed)
3. Use helper methods from `internal/cli/helpers.go` for derived data:
   - `CalculateDuration()` - Calculate work duration from ticket times
   - `ExtractParentID()` - Extract parent ticket ID from Related field
   - `FormatDuration()` - Format duration as human-readable string
4. Build comprehensive JSON response using the returned entity

### Flag Normalization Pattern
Handle both long and short form flags:
```go
type flags struct {
    force       bool
    forceShort  bool
}

func (f *flags) normalize() {
    if f.forceShort {
        f.force = f.forceShort
    }
}
```