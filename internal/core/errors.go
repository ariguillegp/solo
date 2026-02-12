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

// ErrWorktreeDeleteRoot marks attempts to delete the project root worktree.
var ErrWorktreeDeleteRoot = errors.New("cannot delete the project root worktree")

// ErrWorktreeDeleteOutsideRoot marks attempts to delete worktrees outside the managed directory.
var ErrWorktreeDeleteOutsideRoot = errors.New("worktree is outside the managed directory")

// ErrWorktreeUnregistered marks attempts to delete a worktree not registered in git.
var ErrWorktreeUnregistered = errors.New("worktree is not registered")

func IsRecoverableWorktreeDeleteError(err error) bool {
	return errors.Is(err, ErrWorktreeDeleteRoot) ||
		errors.Is(err, ErrWorktreeDeleteOutsideRoot) ||
		errors.Is(err, ErrWorktreeUnregistered)
}
