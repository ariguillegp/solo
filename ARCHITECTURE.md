# ARCHITECTURE

This document explains **how rivet is organized and why**. It is intentionally brief, opinionated, and focused on decision-making constraints rather than an exhaustive file-by-file tour.

## What rivet does

`rivet` is a terminal UI + CLI launcher for project workspaces and tmux-backed tool sessions. It helps you:

1. discover/select a project,
2. discover/create/delete worktrees,
3. launch or switch to a tool session.

The architecture optimizes for predictable behavior, testability, and low coupling between state logic and side effects.

## High-level architecture

Rivet is split into four layers:

- **`internal/core`**: pure domain state machine (model, messages, transitions, filters, effect descriptions).
- **`internal/ui`**: Bubble Tea integration and rendering; converts framework events to core messages and executes effects.
- **`internal/ports`**: interfaces for external capabilities (filesystem/worktrees/sessions).
- **`internal/adapters`**: concrete OS/tmux/git implementations of ports.

`cmd/rv/main.go` wires everything together for interactive and non-interactive entry points.

### Dependency rule

Dependencies point inward:

- `adapters -> ports`
- `ui -> core + ports`
- `core -> (nothing external to domain)`

`core` must stay free of I/O and framework dependencies.

## Core model: source of truth

`internal/core.Model` is the canonical state for user intent and navigation mode:

- current mode (`Browsing`, `Worktree`, `Tool`, `Sessions`, etc.),
- selected project/worktree/tool,
- queries + filtered views,
- warnings/errors,
- warmup/open-session coordination fields.

All user-visible behavior is expressed as:

- **input message** (`Msg*`) + current model
- -> **new model** + list of **effects** (`Eff*`)

This keeps business behavior deterministic and straightforward to test.

## Message/effect loop

Rivet uses an Elm-style loop:

1. UI receives keypress/framework events.
2. UI maps events into `core.Msg`.
3. `core.Update` applies a pure transition and returns `[]core.Effect`.
4. UI effect runner executes effects through ports/adapters.
5. Effect results are converted back into `core.Msg` and fed into `core.Update`.

Important consequence: side effects are **described by core, executed outside core**.

## Mode-driven workflow

The product flow is represented directly as modes:

- `Loading -> Browsing -> Worktree -> Tool -> ToolStarting` for session launch,
- auxiliary confirmation modes for destructive actions,
- `Sessions` mode for tmux attachment,
- `Error` mode for unrecoverable failures.

Mode gates what keys mean, which state fields are editable, and which effects are legal.

## Boundary of responsibility

### `internal/core`

- Domain transitions, filtering, validation and guardrails.
- Describes effects (scan dirs, load worktrees, create/delete, prewarm/open/attach sessions).
- Owns recoverable vs unrecoverable error handling policy.

### `internal/ui`

- Terminal UX composition (Bubble Tea model/view/update).
- Key bindings and view rendering.
- Effect orchestration + asynchronous message delivery.

### `internal/ports`

- Stable contracts for filesystem and session operations.
- Enables test doubles and alternative implementations.

### `internal/adapters`

- OS/git/tmux command execution.
- Path and process concerns.
- Translation of command failures into domain-meaningful errors.

## Invariants and constraints

These are architectural constraints worth preserving:

1. **No I/O in core**: core logic must remain pure and fast to test.
2. **Single source of truth**: avoid duplicating workflow state outside `core.Model`.
3. **Effects are explicit**: if behavior touches external systems, represent it as `Eff*`.
4. **UI orchestrates, core decides**: UI should not embed business rules.
5. **Adapters are replaceable**: business logic should not depend on tmux/git command details.

## Non-interactive path

The CLI flags path (`--project`, `--worktree`, `--tool`, etc.) is intentionally simpler:

- resolve project/worktree/tool,
- create project/worktree when requested/required,
- open or detach tmux session directly.

This path bypasses TUI rendering but still relies on the same domain concepts (`SessionSpec`, supported tools, adapter-backed operations).

## How to extend safely

When adding a feature, prefer this sequence:

1. Add/adjust domain message, transition, and effect in `internal/core`.
2. Add port method(s) only if new external capability is needed.
3. Implement adapter behavior.
4. Update UI mapping/rendering.
5. Add tests at the lowest layer that can express the behavior.

Heuristic: if a change requires shelling out, it should almost never be introduced first in `ui`.

## Trade-offs

- **Pros**
  - High testability of workflow logic.
  - Clear separation between behavior and side effects.
  - Easier refactoring of UI or adapters independently.
- **Cons**
  - More message/effect boilerplate.
  - Some simple changes require touching multiple layers.

These trade-offs are intentional: rivet prioritizes correctness and maintainability of a stateful terminal workflow.
