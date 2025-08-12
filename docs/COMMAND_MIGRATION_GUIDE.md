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

### 2. Commands with App Dependencies

```go
// internal/cli/commands/list.go
type ListCommand struct {
    appFactory func(context.Context) (*cli.App, error)
}

func NewListCommand() command.Command {
    return &ListCommand{
        appFactory: cli.NewApp, // Default factory
    }
}

func (c *ListCommand) Execute(ctx context.Context, flags interface{}, args []string) error {
    app, err := c.appFactory(ctx)
    if err != nil {
        return err
    }
    
    f := flags.(*listFlags)
    return app.ListTickets(ctx, f.status, f.count, f.format)
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

### In Progress 🚧
- [ ] Create migration tickets for remaining commands

### Pending Migration 📋

#### Simple Commands (No Dependencies)

#### Read-Only Commands
- [ ] **status** - Show current ticket status
- [ ] **list** - List tickets with filters
- [ ] **show** - Display ticket details

#### State-Changing Commands
- [ ] **new** - Create new ticket
- [ ] **start** - Start working on ticket
- [ ] **close** - Close current/specified ticket
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