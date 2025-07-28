---
priority: 2
description: Fix golangci-lint errors blocking CI
created_at: "2025-07-28T14:25:15+09:00"
started_at: "2025-07-28T15:33:01+09:00"
closed_at: null
related:
    - parent:250728-001606-fix-worktree-test-macos
---

# Ticket Overview

The CI is failing due to golangci-lint errors. There are currently 26 issues reported:
- 17 errcheck errors (unchecked error returns)
- 9 staticcheck errors (deprecated methods, type inference issues)

These errors are blocking all PRs from passing CI checks.

## Tasks
- [x] Fix errcheck errors in internal/cli/commands.go:283,325,534
- [x] Fix errcheck errors in internal/ticket/manager_test.go:140,145
- [x] Fix errcheck errors in internal/ui/app.go:386,400,402,416,418,427,429,431
- [x] Fix errcheck errors in test/integration/cleanup_test.go:17,18
- [x] Fix errcheck errors in test/integration/directory_creation_test.go:19
- [x] Fix errcheck errors in test/integration/worktree_test.go:181
- [x] Fix staticcheck ST1023 errors in internal/ui/views/new.go:92
- [x] Fix staticcheck SA1019 errors in internal/ui/views/new.go (deprecated Copy() methods)
- [x] Fix staticcheck QF1001 error in internal/ticket/ticket.go:132 (De Morgan's law)
- [x] Run `golangci-lint run` locally to verify all issues are fixed
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Ensure CI passes

## Error Details

### Errcheck Issues (17 total)
```
internal/cli/commands.go:283: Error return value of `app.Git.Checkout` is not checked
internal/cli/commands.go:325: Error return value of `os.Rename` is not checked
internal/cli/commands.go:534: Error return value of `os.Rename` is not checked
internal/ticket/manager_test.go:140: Error return value of `ticket2.Start` is not checked
internal/ticket/manager_test.go:145: Error return value of `os.MkdirAll` is not checked
internal/ui/app.go:386: Error return value of `execCmd.Run` is not checked
internal/ui/app.go:400: Error return value of `m.git.RemoveWorktree` is not checked
internal/ui/app.go:402: Error return value of `m.git.Checkout` is not checked
internal/ui/app.go:416: Error return value of `m.git.RemoveWorktree` is not checked
internal/ui/app.go:418: Error return value of `m.git.Checkout` is not checked
internal/ui/app.go:427: Error return value of `os.Rename` is not checked
internal/ui/app.go:429: Error return value of `m.git.RemoveWorktree` is not checked
internal/ui/app.go:431: Error return value of `m.git.Checkout` is not checked
test/integration/cleanup_test.go:17: Error return value of `os.Chdir` is not checked
test/integration/cleanup_test.go:18: Error return value of `os.Chdir` is not checked
test/integration/directory_creation_test.go:19: Error return value of `os.Chdir` is not checked
test/integration/worktree_test.go:181: Error return value of `os.MkdirAll` is not checked
```

### Staticcheck Issues (9 total)
```
internal/ui/views/new.go:92:11: ST1023: should omit type []tea.Cmd from declaration
internal/ui/views/new.go:214:16: SA1019: styles.FocusedInputStyle.Copy is deprecated
internal/ui/views/new.go:217:16: SA1019: styles.InputStyle.Copy is deprecated
internal/ui/views/new.go:226:19: SA1019: styles.FocusedInputStyle.Copy is deprecated
internal/ui/views/new.go:229:19: SA1019: styles.InputStyle.Copy is deprecated
```

## Notes

These lint errors were discovered while working on PR #9. Some errors were fixed but the CI is catching additional issues. All errors need to be addressed for CI to pass.

## Resolution and Insights

### Summary
Successfully fixed all 27 golangci-lint errors (originally 26, plus 1 discovered during CI):
- 17 errcheck errors for unchecked error returns
- 9 staticcheck errors for deprecated methods and type inference
- 1 additional QF1001 error for De Morgan's law simplification

PR #11 created and CI is now passing.

### Key Patterns Fixed

1. **Error Handling in Rollback Scenarios**
   - Many unchecked errors were in rollback/cleanup code paths
   - Solution: Use `_ = operation()` when the error genuinely can't be handled
   - For critical rollbacks, combine errors in the return message

2. **Test File Error Handling**
   - `os.Chdir()` calls in tests need proper cleanup with defer blocks
   - Pattern: Save original directory, defer restoration with error checking
   ```go
   originalWd, err := os.Getwd()
   require.NoError(t, err)
   defer func() {
       err := os.Chdir(originalWd)
       require.NoError(t, err)
   }()
   ```

3. **Deprecated lipgloss Copy() Method**
   - The `Style.Copy()` method is deprecated in newer versions
   - Solution: Use direct style composition without Copy()
   - Changed from: `InputStyle.Copy().BorderStyle(...)`
   - To: `InputStyle.BorderStyle(...)`

4. **Variable Shadowing in Tests**
   - Multiple `err :=` declarations in same scope cause shadowing
   - Solution: Use `err =` for subsequent assignments

5. **De Morgan's Law Simplification**
   - Complex negated boolean expressions should be simplified
   - Changed from: `!((a && b) || c)` 
   - To: `(!a || !b) && !c`

### CI Considerations
- The CI uses golangci-lint v2.3.0 which may catch different issues than local versions
- Always verify fixes work in the actual worktree, not just the main repo
- macOS test failures due to `/private/var` vs `/var` symlinks are unrelated to lint fixes

### Lessons Learned
1. Run linters in the worktree/branch being worked on for accurate results
2. CI may use different linter versions/configurations than local development
3. Error handling is critical even in cleanup/rollback code paths
4. Keep up with deprecated methods in dependencies to avoid future issues