# rivet AGENTS

Use progressive guidance: start here for orientation, then check `internal/*/AGENTS.md` for details.

## Entry Points
- CLI/TUI entry: `cmd/rv/main.go`
- Interactive mode: `rv [directories...]`
- Non-interactive: `rv --project NAME --worktree BRANCH --tool TOOL [--detach]`

## Architecture Map
- `internal/core`: pure state machine, effects, filtering logic
- `internal/ui`: Bubble Tea model, rendering, effect runner
- `internal/ports`: interfaces for filesystem + sessions
- `internal/adapters`: OS/git/tmux implementations

## Build & Verify
- `make -j validate` (gofmt, golangci-lint, tests, govulncheck)
  - Always use `-j` so test and vulncheck run concurrently after format and lint complete (format → lint → test + vulncheck, avoiding races with `gofmt -w` and `golangci-lint --fix`).
- `make deploy` (install binary)

## Design Guidance
- Keep business logic in `internal/core` (no I/O)
- Keep I/O in `internal/adapters` behind `internal/ports`
- UI should orchestrate effects, not implement core logic
