package core

import "testing"

func TestFilterWorktreesMatchesSanitizedBranch(t *testing.T) {
	worktrees := []Worktree{{Path: "/repo/feat", Name: "solo--feature-test", Branch: "feature/test"}}

	filtered := FilterWorktrees(worktrees, "feature test")
	if len(filtered) != 1 {
		t.Fatalf("expected sanitized branch to match query, got %d", len(filtered))
	}
}
