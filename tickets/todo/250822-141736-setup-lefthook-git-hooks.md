---
priority: 2
description: "Set up Lefthook for automated git hooks to run tests, linting, and formatting on pre-commit and pre-push"
created_at: "2025-08-22T14:17:36+09:00"
started_at: null
closed_at: null
---

# Setup Lefthook Git Hooks

Set up Lefthook as the git hooks manager to automate code quality checks before commits and pushes. This will ensure all code meets quality standards before being committed or pushed to the repository.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [x] Install Lefthook as a development dependency (using Homebrew)
- [x] Create lefthook.yml configuration file with pre-commit hooks for:
  - [x] Running `gofmt` to format code (with stage_fixed)
  - [x] Running `go vet` for static analysis
  - [x] Running `golangci-lint --fast` for linting checks
- [x] Configure pre-push hooks for:
  - [x] Running `make test` to run all tests
  - [x] Running `make build` to verify build
  - [x] Running `make lint` for full linting
- [x] Add lefthook installation to development setup documentation (added `make init`)
- [x] Add .lefthook directory to .gitignore (for local overrides)
- [x] Test the hooks by making a test commit
- [x] Update README.md with git hooks setup instructions
- [x] Update CLAUDE.md with information about the git hooks
- [ ] Get developer approval before closing

## Implementation Details

### Lefthook Configuration
The lefthook.yml file should be structured as:

```yaml
# lefthook.yml
pre-commit:
  parallel: true
  commands:
    fmt:
      run: make fmt
      stage_fixed: true  # Stage files that were fixed
    vet:
      run: make vet
    lint:
      run: make lint
      
pre-push:
  parallel: false  # Run tests sequentially for clear output
  commands:
    test:
      run: make test
```

### Installation Methods
1. **Go install** (recommended for Go projects):
   ```bash
   go install github.com/evilmartians/lefthook@latest
   lefthook install
   ```

2. **As a tool dependency** (tracked in go.mod):
   ```bash
   go get -tool github.com/evilmartians/lefthook
   lefthook install
   ```

### Benefits
- **Performance**: Lefthook is written in Go, making it fast and efficient
- **Parallel execution**: Can run hooks in parallel for better performance
- **Single binary**: No dependencies, works in any environment
- **Flexible configuration**: Supports file filtering, custom scripts, and conditional execution
- **Local overrides**: Developers can customize hooks locally without affecting the team

## Implementation Summary

### What was implemented:
1. **Setup Script** (`scripts/setup-lefthook.sh`):
   - Automatic OS/architecture detection
   - Prioritizes Homebrew installation (as requested)
   - Fallback to go install and direct binary download
   - Colored output for better UX
   - Idempotent (safe to run multiple times)

2. **Lefthook Configuration** (`lefthook.yml`):
   - Pre-commit: Fast checks only (gofmt, go vet, golangci-lint --fast)
   - Pre-push: Comprehensive checks (tests, build, full lint)
   - Parallel execution for pre-commit hooks
   - Sequential execution for pre-push (clearer output)
   - Support for local overrides via `.lefthook-local.yml`

3. **Makefile Integration**:
   - `make setup-hooks`: Runs the setup script
   - `make init`: Complete dev environment setup (deps + hooks + worktree)
   - Updated help text to include new commands

4. **Documentation Updates**:
   - README.md: Added development setup section
   - CLAUDE.md: Added git hooks section with details

### Key Decisions:
- **Homebrew as primary installation method**: Most developers on macOS/Linux have it
- **Tests only on pre-push**: Allows committing work-in-progress without delays
- **Auto-formatting with stage_fixed**: Automatically stages formatted files
- **Skip options documented**: `--no-verify` for emergency commits

## Notes

- Lefthook was chosen over pre-commit because it's written in Go (matching our tech stack) and offers better performance through parallel execution
- The hooks should be non-blocking for developer workflow - failures should provide clear error messages
- Consider adding a skip option for emergencies: `LEFTHOOK_SKIP=1 git commit` or `git commit --no-verify`
- Integration with CI/CD remains separate - hooks provide immediate local feedback while CI provides the authoritative checks