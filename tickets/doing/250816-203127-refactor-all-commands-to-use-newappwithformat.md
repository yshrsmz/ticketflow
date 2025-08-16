---
priority: 2
description: Refactor all CLI commands to use NewAppWithFormat for cleaner initialization
created_at: "2025-08-16T20:31:27+09:00"
started_at: "2025-08-16T22:44:15+09:00"
closed_at: null
related:
    - parent:250816-123703-improve-json-output-separation
---

# Refactor All Commands to Use NewAppWithFormat

## Problem
Currently, many commands still use the old pattern of creating an App then mutating it after flag parsing. This creates awkward state mutations and makes the code harder to understand.

## Current State
Some commands have been updated:
- `cleanup.go` ✅
- `status.go` ✅
- `worktree_list.go` ✅

## Tasks
- [ ] Update `list.go` to use NewAppWithFormat
- [ ] Update `show.go` to use NewAppWithFormat
- [ ] Update `new.go` to use NewAppWithFormat
- [ ] Update `restore.go` to use NewAppWithFormat
- [ ] Update `worktree_clean.go` to use NewAppWithFormat
- [ ] Search for any other commands using old pattern
- [ ] Run `make test` to verify all commands work
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary

## Implementation Pattern
Replace:
```go
app, err := cli.NewApp(ctx)
if err != nil {
    return err
}
// ... parse flags ...
app.Output = cli.NewOutputWriter(os.Stdout, os.Stderr, outputFormat)
app.StatusWriter = cli.NewStatusWriter(os.Stdout, outputFormat)
```

With:
```go
// ... parse flags first ...
outputFormat := cli.ParseOutputFormat(f.format)
app, err := cli.NewAppWithFormat(ctx, outputFormat)
if err != nil {
    return err
}
```

## Benefits
- Cleaner initialization flow
- App is immutable after creation
- No awkward state mutations
- Follows industry best practices (Kong, Docker, kubectl)

## Notes
This is a follow-up to the JSON/text output separation work. The pattern has been proven to work in three commands already.