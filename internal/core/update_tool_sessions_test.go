package core

import "testing"

func TestToolKeyEnterOpensSessionImmediatelyForNoneTool(t *testing.T) {
	m := Model{
		Mode:                 ModeTool,
		SelectedWorktreePath: "/projects/demo/main",
		FilteredTools:        []string{ToolNone},
		ToolIdx:              0,
	}

	updated, effects, handled := UpdateKey(m, KeyEnter)
	if !handled {
		t.Fatalf("expected enter to be handled")
	}
	if updated.Mode != ModeTool {
		t.Fatalf("expected to stay in tool mode, got %v", updated.Mode)
	}
	if updated.PendingSpec != nil {
		t.Fatalf("expected pending spec to stay nil")
	}
	if len(effects) != 1 {
		t.Fatalf("expected one effect, got %d", len(effects))
	}
	eff, ok := effects[0].(EffOpenSession)
	if !ok {
		t.Fatalf("expected EffOpenSession, got %T", effects[0])
	}
	if eff.Spec.DirPath != "/projects/demo/main" || eff.Spec.Tool != ToolNone {
		t.Fatalf("unexpected session spec: %+v", eff.Spec)
	}
}

func TestToolKeyEnterWarmupToolTransitionsToStarting(t *testing.T) {
	m := Model{
		Mode:                 ModeTool,
		SelectedWorktreePath: "/projects/demo/main",
		FilteredTools:        []string{"opencode"},
		ToolIdx:              0,
	}

	updated, effects, handled := UpdateKey(m, KeyEnter)
	if !handled {
		t.Fatalf("expected enter to be handled")
	}
	if updated.Mode != ModeToolStarting {
		t.Fatalf("expected tool-starting mode, got %v", updated.Mode)
	}
	if updated.PendingSpec == nil {
		t.Fatalf("expected pending spec to be set")
	}
	if updated.PendingSpec.Tool != "opencode" || updated.PendingSpec.DirPath != "/projects/demo/main" {
		t.Fatalf("unexpected pending spec: %+v", *updated.PendingSpec)
	}
	if len(effects) != 1 {
		t.Fatalf("expected one effect, got %d", len(effects))
	}
	eff, ok := effects[0].(EffCheckToolReady)
	if !ok {
		t.Fatalf("expected EffCheckToolReady, got %T", effects[0])
	}
	if eff.Spec.Tool != "opencode" || eff.Spec.DirPath != "/projects/demo/main" {
		t.Fatalf("unexpected check-ready spec: %+v", eff.Spec)
	}
}

func TestMsgToolPrewarmExistingOpensPendingSpec(t *testing.T) {
	pending := SessionSpec{DirPath: "/projects/demo/main", Tool: "amp"}
	m := Model{
		Mode:        ModeToolStarting,
		PendingSpec: &pending,
	}

	updated, effects := Update(m, MsgToolPrewarmExisting{Tool: "amp"})
	if updated.PendingSpec != nil {
		t.Fatalf("expected pending spec to be cleared")
	}
	if updated.ToolError != "" {
		t.Fatalf("expected no tool error, got %q", updated.ToolError)
	}
	if len(effects) != 1 {
		t.Fatalf("expected one effect, got %d", len(effects))
	}
	eff, ok := effects[0].(EffOpenSession)
	if !ok {
		t.Fatalf("expected EffOpenSession, got %T", effects[0])
	}
	if eff.Spec != pending {
		t.Fatalf("expected pending spec to be opened, got %+v", eff.Spec)
	}
}

func TestMsgToolDelayElapsedOpensPendingSpec(t *testing.T) {
	pending := SessionSpec{DirPath: "/projects/demo/main", Tool: "codex"}
	m := Model{
		Mode:        ModeToolStarting,
		PendingSpec: &pending,
	}

	updated, effects := Update(m, MsgToolDelayElapsed{Tool: "codex"})
	if updated.PendingSpec != nil {
		t.Fatalf("expected pending spec to be cleared")
	}
	if updated.ToolError != "" {
		t.Fatalf("expected no tool error, got %q", updated.ToolError)
	}
	if len(effects) != 1 {
		t.Fatalf("expected one effect, got %d", len(effects))
	}
	eff, ok := effects[0].(EffOpenSession)
	if !ok {
		t.Fatalf("expected EffOpenSession, got %T", effects[0])
	}
	if eff.Spec != pending {
		t.Fatalf("expected pending spec to be opened, got %+v", eff.Spec)
	}
}

func TestMsgToolPrewarmFailedReturnsToToolMode(t *testing.T) {
	pending := SessionSpec{DirPath: "/projects/demo/main", Tool: "opencode"}
	m := Model{
		Mode:        ModeToolStarting,
		PendingSpec: &pending,
	}

	updated, effects := Update(m, MsgToolPrewarmFailed{Tool: "opencode", Err: errTest("prewarm failed")})
	if updated.Mode != ModeTool {
		t.Fatalf("expected mode tool after failure, got %v", updated.Mode)
	}
	if updated.PendingSpec != nil {
		t.Fatalf("expected pending spec to be cleared")
	}
	if updated.ToolError != "prewarm failed" {
		t.Fatalf("expected tool error to be shown, got %q", updated.ToolError)
	}
	if len(effects) != 0 {
		t.Fatalf("expected no effects, got %d", len(effects))
	}
}

func TestSessionsModeEnterAndLeaveRestoresPreviousMode(t *testing.T) {
	m := Model{
		Mode:             ModeWorktree,
		SessionQuery:     "abc",
		SessionIdx:       3,
		Sessions:         []SessionInfo{{Name: "a"}},
		FilteredSessions: []SessionInfo{{Name: "a"}},
	}

	entered, effects, handled := UpdateKey(m, KeySessions)
	if !handled {
		t.Fatalf("expected sessions key to be handled")
	}
	if entered.Mode != ModeSessions {
		t.Fatalf("expected sessions mode, got %v", entered.Mode)
	}
	if entered.SessionReturnMode != ModeWorktree {
		t.Fatalf("expected return mode worktree, got %v", entered.SessionReturnMode)
	}
	if entered.SessionQuery != "" || entered.SessionIdx != 0 {
		t.Fatalf("expected session state reset, got query=%q idx=%d", entered.SessionQuery, entered.SessionIdx)
	}
	if len(effects) != 1 {
		t.Fatalf("expected one effect, got %d", len(effects))
	}
	if _, ok := effects[0].(EffListSessions); !ok {
		t.Fatalf("expected EffListSessions, got %T", effects[0])
	}

	entered.SessionQuery = "filter"
	entered.SessionIdx = 1
	entered.Sessions = []SessionInfo{{Name: "s1"}}
	entered.FilteredSessions = []SessionInfo{{Name: "s1"}}

	left, leaveEffects, leaveHandled := UpdateKey(entered, KeyBack)
	if !leaveHandled {
		t.Fatalf("expected back key to be handled")
	}
	if left.Mode != ModeWorktree {
		t.Fatalf("expected to return to worktree mode, got %v", left.Mode)
	}
	if left.SessionQuery != "" || left.SessionIdx != 0 {
		t.Fatalf("expected cleared sessions query/index, got query=%q idx=%d", left.SessionQuery, left.SessionIdx)
	}
	if len(left.Sessions) != 0 || len(left.FilteredSessions) != 0 {
		t.Fatalf("expected cleared sessions slices")
	}
	if len(leaveEffects) != 0 {
		t.Fatalf("expected no effects on leave, got %d", len(leaveEffects))
	}
}

func TestSessionsEnterKeyEmitsAttachEffect(t *testing.T) {
	session := SessionInfo{Name: "demo__amp", DirPath: "/projects/demo/main", Tool: "amp"}
	m := Model{
		Mode:             ModeSessions,
		FilteredSessions: []SessionInfo{session},
		SessionIdx:       0,
	}

	updated, effects, handled := UpdateKey(m, KeyEnter)
	if !handled {
		t.Fatalf("expected enter to be handled")
	}
	if updated.Mode != ModeSessions {
		t.Fatalf("expected to stay in sessions mode, got %v", updated.Mode)
	}
	if len(effects) != 1 {
		t.Fatalf("expected one effect, got %d", len(effects))
	}
	eff, ok := effects[0].(EffAttachSession)
	if !ok {
		t.Fatalf("expected EffAttachSession, got %T", effects[0])
	}
	if eff.Session != session {
		t.Fatalf("unexpected session payload: %+v", eff.Session)
	}
}

func TestSessionsNavigationKeysMoveSelection(t *testing.T) {
	m := Model{
		Mode: ModeSessions,
		FilteredSessions: []SessionInfo{
			{Name: "s1"},
			{Name: "s2"},
			{Name: "s3"},
		},
		SessionIdx: 0,
	}

	updated, _, handled := UpdateKey(m, KeyDown)
	if !handled {
		t.Fatalf("expected down key to be handled")
	}
	if updated.SessionIdx != 1 {
		t.Fatalf("expected session index 1, got %d", updated.SessionIdx)
	}

	updated, _, _ = UpdateKey(updated, KeyBottom)
	if updated.SessionIdx != 2 {
		t.Fatalf("expected bottom key to jump to last index, got %d", updated.SessionIdx)
	}
}

func TestInitEmitsScanDirsEffect(t *testing.T) {
	m := NewModel([]string{"/projects"})
	updated, effects := Init(m)
	if updated.Mode != ModeLoading {
		t.Fatalf("expected mode loading, got %v", updated.Mode)
	}
	if len(effects) != 1 {
		t.Fatalf("expected one effect, got %d", len(effects))
	}
	eff, ok := effects[0].(EffScanDirs)
	if !ok {
		t.Fatalf("expected EffScanDirs, got %T", effects[0])
	}
	if len(eff.Roots) != 1 || eff.Roots[0] != "/projects" {
		t.Fatalf("unexpected roots in scan effect: %v", eff.Roots)
	}
}
