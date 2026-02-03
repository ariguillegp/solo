# solo

A lightweight TUI to manage your fleet of agents across all your projects.


## Features

- Fuzzy search across project containers and git worktrees
- Launch tmux sessions running `opencode` or `amp`
- Keyboard-driven navigation
- Delete entire projects (including all worktrees) with confirmation
- Theme picker to customize the UI
- Help bar at the bottom for better discoverability

## Prerequisites

- tmux
- git
- `opencode` and/or `amp`

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
```

Create a new project non-interactively:

```bash
solo --project my-project --worktree main --tool opencode --create-project
```

### Theme Picker

Press `ctrl+t` to open the theme picker and choose from the available themes. The
selected theme updates UI colors across the app.

### Project/Worktree Deletion

Press `ctrl+d` while browsing projects to delete the selected project and all of its
worktrees. Similarly you can use `ctrl+d` to delete worktrees in the worktree selection
view. `solo` asks for confirmation before performing the deletion.

### Project Layout

Solo expects projects to be valid git repositories so additional worktrees can be created under `~/.solo/worktrees/`

Stale worktree references (from manually deleted directories) are automatically
pruned whenever the worktree list is loaded, keeping the list accurate.

## License

MIT
