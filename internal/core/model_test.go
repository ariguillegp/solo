package core

import "testing"

func TestCreateWorktreeNameRejectsSanitizedDuplicate(t *testing.T) {
	m := Model{
		WorktreeQuery: "feature test",
		Worktrees:     []Worktree{{Branch: "feature/test"}},
	}

	if name, ok := m.CreateWorktreeName(); ok {
		t.Fatalf("expected duplicate branch to be rejected, got %q", name)
	}
}
