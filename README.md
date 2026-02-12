# rivet

A lightweight TUI to manage your fleet of agents across all your projects.

## Main Features at a Glance

- Guided 3-step workflow: pick a project, pick/create a workspace (git worktree), then launch a tool (`opencode`, `amp`, `claude`, `codex`, or `none`).
- Fast fuzzy filtering in every step (projects, workspaces, tools, and sessions).
- Built-in tmux session switcher: press `ctrl+s` from the main screens to open **Active tmux sessions**, filter them, and press `enter` to attach.
- Warm-start support for tool sessions, plus session reuse/caching for quicker reopen flows.
- Project/workspace lifecycle management in-app (create and delete with confirmation and cleanup).
- Stale worktree references (from manually deleted directories) are automatically pruned whenever the worktree list is loaded, keeping the list accurate.
- Keyboard-first UX with help modal (`?`) and theme picker (`ctrl+t`).
- Optional non-interactive mode for launching sessions directly via CLI flags.
- Theme picker to customize the UI
- Help bar at the bottom for better discoverability

## Run agent tool in worktree (session caching)
After selecting a project/worktree tuple, the program automatically opens new sessions for tools that benefit from warm starts.

https://github.com/user-attachments/assets/d0d314d5-413f-4480-b6aa-0523587ff8cc

## Create/Delete worktree
Deleting a worktree also kills any session using it.

https://github.com/user-attachments/assets/39cfb6ba-c9df-4652-a9ba-b4ef3fe3aeeb

## Create/Delete project
Deleting a project also kills any sessions using it.

https://github.com/user-attachments/assets/bc3ebf0f-1152-40af-98df-d198e5638302

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

Add the following line to your `~/.tmux.conf` so you can use `tmux-prefix + f` to launch `rv` from a tmux session

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

Add the following line to your `~/.tmux.conf` so you can use `tmux-prefix + f` to launch `rv` from a tmux session

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

### Non-Interactive Launch

Open a session directly without the UI:

```bash
rv --project my-project --worktree main --tool opencode [--detach]

rv --project my-project --worktree main --tool claude [--detach]

rv --project my-project --worktree main --tool codex [--detach]

rv --project my-project --worktree main --tool none [--detach]
```

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
