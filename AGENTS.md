# Repository Guidelines

## Project Structure & Modules
- `cmd/ticketflow/` – CLI/TUI entrypoint (`main.go`).
- `internal/` – core packages: `config/`, `ticket/`, `git/`, `cli/`, `ui/`, `log/`, `worktree/`.
- `test/integration/` – black‑box integration tests; additional `*_integration_test.go` live under `internal/cli/commands/`.
- `tickets/` – example tickets (`todo/`, `doing/`, `done/`).
- `docs/`, `benchmarks/`, `scripts/` – documentation, perf, and helper scripts.

## Build, Test, and Development
- `make init` – download deps, set up git hooks and worktree helpers.
- `make build` / `make run` – compile and run `ticketflow` locally.
- `make test` – run all tests; `make test-unit`, `make test-integration` for subsets.
- `make coverage` – generate `coverage.html`.
- `make fmt` `make vet` `make lint` – format, vet, and (optionally) run `golangci-lint`.
- `make build-all` – cross‑compile into `dist/`; `make version` prints ldflags.

## Coding Style & Naming
- Go 1.24.x (see `mise.toml`). Always `make fmt` and fix `vet` findings.
- Prefer small packages under `internal/`; avoid new deps unless necessary.
- Use contexts for operations that touch git or I/O; honor cancellation.
- Log via `internal/log` (not `fmt`) for non‑UI diagnostics.
- Ticket files: `tickets/<status>/<YYMMDD-HHMMSS>-<slug>.md` with YAML frontmatter.

## Testing Guidelines
- Frameworks: Go testing + Testify. Use table‑driven tests.
- Unit tests co‑locate as `*_test.go`; integration tests live under `test/integration/` or `internal/cli/commands/*_integration_test.go`.
- Keep tests self‑contained (no network). Use `internal/testutil` and temporary repos when interacting with git.
- Run `make test` locally; add/adjust tests for any behavior change.

## Commit & Pull Requests
- Commit messages: imperative mood, include ticket ID when applicable.
  - Example: `start: improve worktree init messaging (250916-235037)`
- PRs must include: what/why, test plan (commands run), and CLI/TUI before/after if UX changes.
- Link the related ticket (e.g., `tickets/doing/...`) and ensure `make test` passes.
- Keep diffs focused; avoid drive‑by refactors.

## Agent-Specific Notes
- On initialization, read `CLAUDE.md` (repo root) and follow its guidance alongside this document; re-check it when switching worktrees.
- Prefer minimal, targeted patches; align with existing patterns and naming.
- Update README or command help when modifying CLI/TUI behavior.
- Don’t commit binaries or `dist/`; don’t add license headers.

### Execution & Approvals (Required)
- Run the commands you need (tests, builds, fmt, vet, lint, etc.) as part of normal development – don’t skip critical validation steps.
- When a command needs elevated permissions, touches outside the workspace, triggers network access, or might surprise the user, explain why and request approval instead of silently skipping it.
- Give a concise preview of grouped commands before running them (e.g., “Run fmt, vet, then unit+integration tests”).

### Go Toolchain for Tests/Builds
- The project targets Go 1.24.x (see `mise.toml`). When running tests/builds, use the matching toolchain.
- If a toolchain mismatch occurs, ask for approval to run with the correct toolchain, for example:
  - `GOTOOLCHAIN=go1.24.6 make test`
- If a tool installer (e.g., `mise`) must be used, request permission before invoking it since it may require network access.
