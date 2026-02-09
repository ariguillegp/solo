package core

import "errors"

var ErrWorktreeExists = errors.New("worktree already exists")
var ErrInvalidWorktreeName = errors.New("invalid worktree name")
