<p align="center">
  <img src="assets/logo.png" width="200" alt="rivet logo">
</p>

## Description
A lightweight TUI to manage your fleet of agents across all your projects.

## Main Features at a Glance

- Guided 3-step workflow: pick a project, pick/create a workspace (git worktree), then launch a tool (`opencode`, `amp`, `claude`, `codex`, or `none`).
- Fast fuzzy filtering in every step (projects, workspaces, tools, and sessions).
- Scans `~/Projects` (or provided roots) up to 2 levels deep, skipping hidden/common vendor directories.
- Built-in tmux session switcher: press `ctrl+s` from the main screens to open **Active tmux sessions**, filter them, and press `enter` to attach.
- Tool sessions are prewarmed in the background and reused if already running (warm-starts for `opencode`, `amp`, `claude`, and `codex`; `none` opens a shell).
- Project/workspace lifecycle management in-app (create and delete with confirmation and cleanup). Worktree deletions are limited to rivet-managed worktrees under `~/.rivet/worktrees` (project root is protected).
- Stale worktree references (from manually deleted directories) are automatically pruned whenever the worktree list is loaded, keeping the list accurate.
- Keyboard-first UX with help modal (`?`), theme picker (`ctrl+t`), and a persistent help bar.
- Optional non-interactive mode for launching sessions directly via CLI flags.

## Run agent tool in worktree (session caching)
After selecting a project/worktree tuple, the program prewarms all supported tools in the background (creating tmux sessions if needed). Existing sessions are reused, and selecting `none` opens a shell immediately.

https://github.com/user-attachments/assets/e1905edd-01ad-441b-b12e-38f3e2395645

## Create/Delete worktree
Deleting a worktree also kills any session using it. Only the project root and rivet-managed worktrees under `~/.rivet/worktrees` are listed, and the root worktree cannot be deleted from the UI.

https://github.com/user-attachments/assets/a6b2735a-20b2-49c9-ad0b-47e9e7349bdb

## Create/Delete project
Deleting a project also kills any sessions using it.

https://github.com/user-attachments/assets/cefca0ef-3b09-402b-904a-78ca328d43a6

## Help Menu
Common actions are visible in the help bar at the bottom for better discoverability, but a more comprehensive help menu is one key press "?" away on every view.

https://github.com/user-attachments/assets/e700db3d-d511-4ef9-886f-ad007691709d

## Prerequisites

- tmux
- git
- `opencode`, `amp`, `claude`, and/or `codex` (optional for `none` sessions)
- Projects must be valid git repositories. The tool by default will look for projects under `~/Projects` and additional worktrees will be created under `~/.rivet/worktrees/`

## Installation

```bash
$ git clone git@github.com:ariguillegp/rivet.git
$ cd rivet
$ make install
```

## Usage

### Recommended way

Create keybindings to run this tool from your regular shell environment and from inside tmux sessions. If are you are not using `~/Projects/` as your base directory for your project repositories, you will need to run `rv YOUR_BASE_DIR` to find out the repos you wanna work on.

**Bash**

Add the following line to your `~/.bashrc`

```bash
bind -x '"\C-f": "rv YOUR_BASE_DIR"'
```

**tmux**

Add the following line to your `~/.config/tmux/tmux.conf` so you can use `tmux-prefix + f` to launch `rv` from a tmux session

```tmux
bind-key f run-shell "tmux has-session -t rv-launcher 2>/dev/null && tmux kill-session -t rv-launcher; tmux new-session -d -s rv-launcher 'bash -lc \"rv YOUR_BASE_DIR\"'; tmux switch-client -t rv-launcher"
```

This launches rv in a temporary tmux session to keep your current session clean.

**Zsh**

Add the following line to your `~/.zshrc`

```bash
bindkey -s '^f' 'rv YOUR_BASE_DIR\n'
```

**tmux**

Add the following line to your `~/.config/tmux/tmux.conf` so you can use `tmux-prefix + f` to launch `rv` from a tmux session

```tmux
bind-key f run-shell "tmux has-session -t rv-launcher 2>/dev/null && tmux kill-session -t rv-launcher; tmux new-session -d -s rv-launcher 'zsh -lc \"rv YOUR_BASE_DIR\"'; tmux switch-client -t rv-launcher"
```

This launches rv in a temporary tmux session to keep your current session clean.

### Interactive launch

```bash
rv [directories...]
```

Rivet starts tools directly inside tmux sessions using your default shell, so no
login shell flags are required.

By default, rv scans `~/Projects` (personal preference). Pass custom directories as arguments:

```bash
rv ~/projects ~/work
```

Scanning goes up to 2 directory levels deep and skips common vendor/cache directories such as `.git`, `node_modules`, `vendor`, `.cache`, `.venv`, `__pycache__`, and `target`.

### Non-Interactive Launch

Open a session directly without the UI:

```bash
rv --project my-project --worktree main --tool opencode [--detach]

rv --project my-project --worktree main --tool claude [--detach]

rv --project my-project --worktree main --tool codex [--detach]

rv --project my-project --worktree main --tool none [--detach]
```

`--project` and `--worktree` accept names or paths. If the worktree doesn't exist yet, rivet creates a new worktree/branch automatically. Use `--create-project` to initialize a missing project (in the first root or at the provided path).

Create a new project non-interactively:

```bash
rv --project my-project --worktree main --tool opencode --create-project
```

## Acknowledgments

Inspired by:
* [agent-of-empires](https://github.com/njbrake/agent-of-empires) (Rust + ratatui + tmux)
* [agent-deck](https://github.com/asheshgoplani/agent-deck) (GO + BubbleTea + tmux)

## License

MIT
