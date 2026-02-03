#!/usr/bin/env bash
set -euo pipefail

go install github.com/ariguillegp/solo/cmd/solo@latest
mkdir -p "$HOME/Projects"
mkdir -p "$HOME/.solo/worktrees"
