#!/usr/bin/env bash
# Source this file in your .bashrc or .zshrc
# Add: source /path/to/solo.sh

SOLO_BIN="${SOLO_BIN:-solo}"

_solo_widget() {
    "$SOLO_BIN"
}

# Zsh
if [[ -n "$ZSH_VERSION" ]]; then
    zle -N _solo_widget
    bindkey '^f' _solo_widget
# Bash
elif [[ -n "$BASH_VERSION" ]]; then
    bind -x '"\C-f": _solo_widget'
fi
