# solo

A lightweight TUI to manage your fleet of agents across all your projects.

https://github.com/user-attachments/assets/be955ba5-ba2e-4ff0-9047-e9b610c39c34

## Features

- Fuzzy search across project containers and git worktrees
- Launch tmux sessions running `opencode` or `amp`
- Keyboard-driven navigation
- Theme picker to customize the UI
- Help bar at the bottom for better discoverability

## Prerequisites

- tmux
- git
- `opencode` and/or `amp`

## Installation

```bash
go install github.com/ariguillegp/solo/cmd/solo@latest
```

## Usage

### Recommended way

Create keybindings to run this tool from your regular shell environment and from inside a tmux session

**Bash**

Add the following line to your `~/.bashrc`

```bash
bind -x '"\C-f": "solo"'
```

**Zsh**

Add the following line to your `~/.zshrc`

```bash
bindkey -s '^f' 'solo\n'
```

**tmux**

Add the following line to your `~/.tmux.conf`

```tmux
bind-key f run-shell "tmux has-session -t solo-launcher 2>/dev/null && tmux kill-session -t solo-launcher; tmux new-session -d -s solo-launcher 'bash -lc \"solo\"'; tmux switch-client -t solo-launcher"
```

This launches solo in a temporary tmux session to keep your current session clean.

### Interactive launch

```bash
solo [directories...]
```

Solo starts tools directly inside tmux sessions using your default shell, so no
login shell flags are required.

By default, solo scans `~/Work/tries` (personal preference). Pass custom directories as arguments:

```bash
solo ~/projects ~/work
```

Solo scans up to two directory levels under each root and skips hidden or ignored
directories (like `.git`, `.bare`, `node_modules`, and `vendor`).

### Non-Interactive Launch

Open a session directly without the UI:

```bash
solo --project my-project --worktree main --tool opencode [--detach]
```

`--project` can be a container name or a path (absolute or relative). `--worktree` can be
an existing worktree path (which must already exist) or a branch name; if a named worktree
does not exist, solo creates it.

Create a new project non-interactively:

```bash
solo --project my-project --worktree main --tool opencode --create-project
```

### Theme Picker

Press `ctrl+t` to open the theme picker and choose from the available themes. The
selected theme updates UI colors across the app.

### Project Layout

Solo expects a project container that holds one worktree per directory:

```
my-project/
  .bare/
  main/
  feat-login/
  fix-bug/
```

The project directory itself stores a bare repo in `.bare`, while each worktree is a
separate directory. The `main` worktree is created without an initial commit.

Stale worktree references (from manually deleted directories) are automatically
pruned whenever the worktree list is loaded, keeping the list accurate.

## License

MIT
