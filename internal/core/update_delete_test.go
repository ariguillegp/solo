package core

import "testing"

func TestWorktreeDeleteKeyEntersConfirm(t *testing.T) {
	m := Model{
		Mode:            ModeWorktree,
		SelectedProject: "/projects/demo",
		FilteredWT: []Worktree{
			{Path: "/projects/demo/feature", Name: "feature", Branch: "feature"},
		},
		WorktreeIdx: 0,
	}

	updated, effects, handled := UpdateKey(m, "ctrl+d")
	if !handled {
		t.Fatal("expected ctrl+d to be handled")
	}
	if updated.Mode != ModeWorktreeDeleteConfirm {
		t.Fatalf("expected delete confirm mode, got %v", updated.Mode)
	}
	if updated.WorktreeDeletePath != "/projects/demo/feature" {
		t.Fatalf("expected delete path to be set, got %q", updated.WorktreeDeletePath)
	}
	if len(effects) != 0 {
		t.Fatalf("expected no effects, got %d", len(effects))
	}
}

func TestProjectDeleteKeyEntersConfirm(t *testing.T) {
	m := Model{
		Mode: ModeBrowsing,
		Filtered: []DirEntry{
			{Path: "/projects/demo", Name: "demo", Exists: true},
		},
		SelectedIdx: 0,
	}

	updated, effects, handled := UpdateKey(m, "ctrl+d")
	if !handled {
		t.Fatal("expected ctrl+d to be handled")
	}
	if updated.Mode != ModeProjectDeleteConfirm {
		t.Fatalf("expected delete confirm mode, got %v", updated.Mode)
	}
	if updated.ProjectDeletePath != "/projects/demo" {
		t.Fatalf("expected delete path to be set, got %q", updated.ProjectDeletePath)
	}
	if len(effects) != 0 {
		t.Fatalf("expected no effects, got %d", len(effects))
	}
}

func TestWorktreeDeleteKeyWithNoSelectionNoop(t *testing.T) {
	m := Model{Mode: ModeWorktree}

	updated, effects, handled := UpdateKey(m, "ctrl+d")
	if !handled {
		t.Fatal("expected ctrl+d to be handled")
	}
	if updated.Mode != ModeWorktree {
		t.Fatalf("expected to stay in worktree mode, got %v", updated.Mode)
	}
	if updated.WorktreeDeletePath != "" {
		t.Fatalf("expected delete path to stay empty, got %q", updated.WorktreeDeletePath)
	}
	if len(effects) != 0 {
		t.Fatalf("expected no effects, got %d", len(effects))
	}
}

func TestWorktreeDeleteConfirmEnterEmitsEffect(t *testing.T) {
	m := Model{
		Mode:               ModeWorktreeDeleteConfirm,
		SelectedProject:    "/projects/demo",
		WorktreeDeletePath: "/projects/demo/feature",
	}

	updated, effects, handled := UpdateKey(m, "enter")
	if !handled {
		t.Fatal("expected enter to be handled")
	}
	if updated.Mode != ModeWorktreeDeleteConfirm {
		t.Fatalf("expected to stay in confirm mode, got %v", updated.Mode)
	}
	if len(effects) != 1 {
		t.Fatalf("expected one effect, got %d", len(effects))
	}
	eff, ok := effects[0].(EffDeleteWorktree)
	if !ok {
		t.Fatalf("expected EffDeleteWorktree, got %T", effects[0])
	}
	if eff.ProjectPath != "/projects/demo" || eff.WorktreePath != "/projects/demo/feature" {
		t.Fatalf("unexpected effect payload: %+v", eff)
	}
}

func TestProjectDeleteConfirmEnterEmitsEffect(t *testing.T) {
	m := Model{
		Mode:              ModeProjectDeleteConfirm,
		ProjectDeletePath: "/projects/demo",
	}

	updated, effects, handled := UpdateKey(m, "enter")
	if !handled {
		t.Fatal("expected enter to be handled")
	}
	if updated.Mode != ModeProjectDeleteConfirm {
		t.Fatalf("expected to stay in confirm mode, got %v", updated.Mode)
	}
	if len(effects) != 1 {
		t.Fatalf("expected one effect, got %d", len(effects))
	}
	eff, ok := effects[0].(EffDeleteProject)
	if !ok {
		t.Fatalf("expected EffDeleteProject, got %T", effects[0])
	}
	if eff.ProjectPath != "/projects/demo" {
		t.Fatalf("unexpected effect payload: %+v", eff)
	}
}

func TestWorktreeDeleteConfirmEscCancels(t *testing.T) {
	m := Model{
		Mode:               ModeWorktreeDeleteConfirm,
		WorktreeDeletePath: "/projects/demo/feature",
	}

	updated, effects, handled := UpdateKey(m, "esc")
	if !handled {
		t.Fatal("expected esc to be handled")
	}
	if updated.Mode != ModeWorktree {
		t.Fatalf("expected to return to worktree mode, got %v", updated.Mode)
	}
	if updated.WorktreeDeletePath != "" {
		t.Fatalf("expected delete path to be cleared, got %q", updated.WorktreeDeletePath)
	}
	if len(effects) != 0 {
		t.Fatalf("expected no effects, got %d", len(effects))
	}
}

func TestProjectDeleteConfirmEscCancels(t *testing.T) {
	m := Model{
		Mode:              ModeProjectDeleteConfirm,
		ProjectDeletePath: "/projects/demo",
	}

	updated, effects, handled := UpdateKey(m, "esc")
	if !handled {
		t.Fatal("expected esc to be handled")
	}
	if updated.Mode != ModeBrowsing {
		t.Fatalf("expected to return to browsing mode, got %v", updated.Mode)
	}
	if updated.ProjectDeletePath != "" {
		t.Fatalf("expected delete path to be cleared, got %q", updated.ProjectDeletePath)
	}
	if len(effects) != 0 {
		t.Fatalf("expected no effects, got %d", len(effects))
	}
}

func TestMsgWorktreeDeletedErrorSetsModeError(t *testing.T) {
	base := Model{Mode: ModeWorktreeDeleteConfirm}

	updated, effects := Update(base, MsgWorktreeDeleted{Err: errTest("boom")})
	if updated.Mode != ModeError {
		t.Fatalf("expected error mode, got %v", updated.Mode)
	}
	if updated.Err == nil || updated.Err.Error() != "boom" {
		t.Fatalf("expected error to be set, got %v", updated.Err)
	}
	if len(effects) != 0 {
		t.Fatalf("expected no effects, got %d", len(effects))
	}
}

func TestMsgWorktreeDeletedReloadsWorktrees(t *testing.T) {
	base := Model{SelectedProject: "/projects/demo"}

	updated, effects := Update(base, MsgWorktreeDeleted{Path: "/projects/demo/feature"})
	if updated.Mode != ModeWorktree {
		t.Fatalf("expected worktree mode, got %v", updated.Mode)
	}
	if updated.WorktreeDeletePath != "" {
		t.Fatalf("expected delete path to be cleared, got %q", updated.WorktreeDeletePath)
	}
	if len(effects) != 1 {
		t.Fatalf("expected one effect, got %d", len(effects))
	}
	eff, ok := effects[0].(EffLoadWorktrees)
	if !ok {
		t.Fatalf("expected EffLoadWorktrees, got %T", effects[0])
	}
	if eff.ProjectPath != "/projects/demo" {
		t.Fatalf("unexpected project path: %q", eff.ProjectPath)
	}
}

func TestMsgProjectDeletedReloadsProjects(t *testing.T) {
	base := Model{RootPaths: []string{"/projects"}, Mode: ModeProjectDeleteConfirm}

	updated, effects := Update(base, MsgProjectDeleted{ProjectPath: "/projects/demo"})
	if updated.Mode != ModeBrowsing {
		t.Fatalf("expected browsing mode, got %v", updated.Mode)
	}
	if updated.ProjectDeletePath != "" {
		t.Fatalf("expected delete path to be cleared, got %q", updated.ProjectDeletePath)
	}
	if len(effects) != 1 {
		t.Fatalf("expected one effect, got %d", len(effects))
	}
	eff, ok := effects[0].(EffScanDirs)
	if !ok {
		t.Fatalf("expected EffScanDirs, got %T", effects[0])
	}
	if len(eff.Roots) != 1 || eff.Roots[0] != "/projects" {
		t.Fatalf("unexpected roots: %v", eff.Roots)
	}
}

type errTest string

func (e errTest) Error() string {
	return string(e)
}
