package core

import "errors"

// ErrWorktreeExists marks errors caused by trying to create a worktree for an existing branch.
var ErrWorktreeExists = errors.New("worktree already exists")

// WorktreeExistsError captures the branch name when a duplicate is detected.
type WorktreeExistsError struct {
	Branch string
}

func (e WorktreeExistsError) Error() string {
	if e.Branch == "" {
		return ErrWorktreeExists.Error()
	}
	return "worktree already exists for branch " + e.Branch
}

func (e WorktreeExistsError) Is(target error) bool {
	return target == ErrWorktreeExists
}

func IsWorktreeExistsError(err error) bool {
	return errors.Is(err, ErrWorktreeExists)
}
