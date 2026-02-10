# solo

A lightweight TUI to manage your fleet of agents across all your projects.

## Run agent tool in worktree (session caching)
After selecting a project/worktree tuple, the program automatically opens new sessions for tools that benefit from warm starts.

https://github.com/user-attachments/assets/5d8f25d3-4344-4801-ab51-fb3d00b22c23

## Create/Delete worktree
Deleting a worktree also kills any session using it.

https://github.com/user-attachments/assets/1d38e6ff-1a80-4322-802b-6f117ae1a05f

## Create/Delete project
Deleting a project also kills any sessions using it.

https://github.com/user-attachments/assets/4d98a9ac-4e05-4ca3-857e-e670c1a9d65a

## Features

- Fuzzy search across project containers and git worktrees
- Launch tmux sessions running `opencode`, `amp`, `claude`, `codex`, or `none`
- Keyboard-driven navigation
- Delete entire projects (including all worktrees) with confirmation
- Theme picker to customize the UI
- Help bar at the bottom for better discoverability

## Prerequisites

- tmux
- git
- `opencode`, `amp`, `claude`, and/or `codex` (optional for `none` sessions)

## Installation

```bash
$ git clone git@github.com:ariguillegp/solo.git
$ cd solo
$ make install
```

## Usage

### Recommended way

Create keybindings to run this tool from your regular shell environment and from inside tmux sessions. If are you are not using `~/Projects/` as your base directory for your project repositories, you will need to run `solo YOUR_BASE_DIR` to find out the repos you wanna work on.

**Bash**

Add the following line to your `~/.bashrc`

```bash
bind -x '"\C-f": "solo YOUR_BASE_DIR"'
```

**tmux**

Add the following line to your `~/.tmux.conf` so you can use `tmux-prefix + f` to launch `solo` from a tmux session

```tmux
bind-key f run-shell "tmux has-session -t solo-launcher 2>/dev/null && tmux kill-session -t solo-launcher; tmux new-session -d -s solo-launcher 'bash -lc \"solo YOUR_BASE_DIR\"'; tmux switch-client -t solo-launcher"
```

This launches solo in a temporary tmux session to keep your current session clean.

**Zsh**

Add the following line to your `~/.zshrc`

```bash
bindkey -s '^f' 'solo YOUR_BASE_DIR\n'
```

**tmux**

Add the following line to your `~/.tmux.conf` so you can use `tmux-prefix + f` to launch `solo` from a tmux session

```tmux
bind-key f run-shell "tmux has-session -t solo-launcher 2>/dev/null && tmux kill-session -t solo-launcher; tmux new-session -d -s solo-launcher 'zsh -lc \"solo YOUR_BASE_DIR\"'; tmux switch-client -t solo-launcher"
```

This launches solo in a temporary tmux session to keep your current session clean.

### Interactive launch

```bash
solo [directories...]
```

Solo starts tools directly inside tmux sessions using your default shell, so no
login shell flags are required.

By default, solo scans `~/Projects` (personal preference). Pass custom directories as arguments:

```bash
solo ~/projects ~/work
```

### Non-Interactive Launch

Open a session directly without the UI:

```bash
solo --project my-project --worktree main --tool opencode [--detach]

solo --project my-project --worktree main --tool claude [--detach]

solo --project my-project --worktree main --tool codex [--detach]

solo --project my-project --worktree main --tool none [--detach]
```

Create a new project non-interactively:

```bash
solo --project my-project --worktree main --tool opencode --create-project
```

### Theme Picker

Press `ctrl+t` to open the theme picker and choose from the available themes. The
selected theme updates UI colors across the app.

### Project Layout

Solo expects projects to be valid git repositories, which by default will be found under `~/Projects` and additional worktrees will be created under `~/.solo/worktrees/`

Stale worktree references (from manually deleted directories) are automatically
pruned whenever the worktree list is loaded, keeping the list accurate.

## Acknowledgments

Inspired by:
* [agent-of-empires](https://github.com/njbrake/agent-of-empires) (Rust + ratatui + tmux)
* [agent-deck](https://github.com/asheshgoplani/agent-deck) (GO + BubbleTea + tmux)

## License

MIT
