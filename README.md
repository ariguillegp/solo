# solo

A fast terminal-based project picker that launches tmux sessions for agent tools.

## Features

- Fuzzy search across project containers (directories with git worktrees)
- Git worktree selection and branch creation
- Create new projects with a default `main` worktree and empty commit
- Launch tmux sessions running `opencode` or `amp`
- Keyboard-driven navigation

## Prerequisites

- tmux
- git
- `opencode` and/or `amp` on your PATH

## Installation

```bash
go install github.com/ariguillegp/solo/cmd/solo@latest
```

## Usage

```bash
solo [directories...]
```

Solo launches tmux sessions by running your shell as a login shell with `-l`
(e.g., `bash -lc`). Supported examples include `bash` and `zsh`. Shells such as
`fish` use `--login` instead. If your login shell does not accept `-l`, set your
`SHELL` to a compatible shell before using solo.

By default, solo scans `~/Work/tries`. Pass custom directories as arguments:

```bash
solo ~/projects ~/work
```

### Non-Interactive Launch

Open a session directly without the UI:

```bash
solo --project my-project --worktree main --tool opencode
```

`--project` can be a container name or an absolute path. `--worktree` can be an existing
worktree directory or a branch name; if it does not exist, solo creates it.

Create a new project non-interactively:

```bash
solo --project my-project --worktree main --tool opencode --create-project
```

### Configurable Tools

By default, solo offers `opencode` and `amp`. You can override the list with the
`SOLO_TOOLS` environment variable (comma or whitespace separated):

```bash
SOLO_TOOLS="amp,opencode,custom-agent" solo
```

This affects both the UI tool picker and `--tool` validation.

To create the tmux session without attaching (useful for scripts), add `--detach`.

### Project Layout

Solo expects a project container that holds one worktree per directory:

```
my-project/
  main/
  feat-login/
  fix-bug/
```

The project directory itself is not a git repo; the `main` worktree is the primary repo.

### Keybinding Examples

Optional bindings to launch `solo` quickly:

**Bash**

```bash
bind -x '"\C-f": "solo"'
```

**Zsh**

```bash
bindkey -s '^f' 'solo\n'
```

**tmux**

```tmux
bind-key f run-shell "tmux has-session -t solo-launcher 2>/dev/null && tmux kill-session -t solo-launcher; tmux new-session -d -s solo-launcher 'bash -lc \"solo\"'; tmux switch-client -t solo-launcher"
```

This launches solo in a temporary tmux session to keep your current session clean.

### Workflow

1. Launch solo and type to fuzzy-filter project containers
2. Press `enter` to select a project (or create a new one)
3. Choose an existing worktree or type a branch name to create a new one
4. Select a tool (`opencode` or `amp`) to start the session

### Keybindings

#### Project Selection

| Key | Action |
|-----|--------|
| `↑` / `ctrl+k` | Previous suggestion |
| `↓` / `ctrl+j` | Next suggestion |
| `enter` | Select project (go to worktree selection) |
| `ctrl+n` | Create new project |
| `esc` / `ctrl+c` | Quit |

#### Worktree Selection

| Key | Action |
|-----|--------|
| `↑` / `ctrl+k` | Previous worktree |
| `↓` / `ctrl+j` | Next worktree |
| `enter` | Select worktree / create new if typing |
| `ctrl+n` | Create new worktree with typed branch name |
| `esc` | Go back to project selection |
| `ctrl+c` | Quit |

#### Tool Selection

| Key | Action |
|-----|--------|
| `↑` / `ctrl+k` | Previous tool |
| `↓` / `ctrl+j` | Next tool |
| `enter` | Start session |
| `esc` | Back to worktree selection |
| `ctrl+c` | Quit |

## License

MIT
