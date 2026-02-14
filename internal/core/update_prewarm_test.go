package core

import (
	"testing"
	"time"
)

func TestEnterToolModeInitializesWarmupCounts(t *testing.T) {
	m := Model{
		Mode:            ModeWorktree,
		SelectedProject: "/projects/demo",
		Tools:           []string{"opencode", "claude", ToolNone},
	}

	updated, effects := Update(m, MsgWorktreeCreated{Path: "/projects/demo/wt"})
	if updated.Mode != ModeTool {
		t.Fatalf("expected tool mode, got %v", updated.Mode)
	}
	if updated.ToolWarmupTotal != 2 {
		t.Fatalf("expected 2 tools needing warmup, got %d", updated.ToolWarmupTotal)
	}
	if updated.ToolWarmupCompleted != 0 {
		t.Fatalf("expected completed count to start at 0, got %d", updated.ToolWarmupCompleted)
	}
	if updated.ToolWarmupFailed != 0 {
		t.Fatalf("expected failed count to start at 0, got %d", updated.ToolWarmupFailed)
	}

	if len(effects) != 1 {
		t.Fatalf("expected one effect, got %d", len(effects))
	}
	eff, ok := effects[0].(EffPrewarmAllTools)
	if !ok {
		t.Fatalf("expected EffPrewarmAllTools, got %T", effects[0])
	}
	if len(eff.Tools) != 2 {
		t.Fatalf("expected 2 warmup tools in effect, got %d", len(eff.Tools))
	}
}

func TestPrewarmMessagesUpdateWarmupCounters(t *testing.T) {
	base := Model{
		Mode:            ModeTool,
		ToolWarmupTotal: 3,
		ToolWarmStart:   map[string]time.Time{},
		ToolErrors:      map[string]string{},
	}

	updated, _ := Update(base, MsgToolPrewarmStarted{Tool: "claude", StartedAt: time.Now()})
	if updated.ToolWarmupCompleted != 1 {
		t.Fatalf("expected completed count to be 1 after started message, got %d", updated.ToolWarmupCompleted)
	}
	if updated.ToolWarmupFailed != 0 {
		t.Fatalf("expected failed count to remain 0, got %d", updated.ToolWarmupFailed)
	}

	updated, _ = Update(updated, MsgToolPrewarmExisting{Tool: "opencode"})
	if updated.ToolWarmupCompleted != 2 {
		t.Fatalf("expected completed count to be 2 after existing message, got %d", updated.ToolWarmupCompleted)
	}
	if updated.ToolWarmupFailed != 0 {
		t.Fatalf("expected failed count to remain 0 after existing message, got %d", updated.ToolWarmupFailed)
	}

	updated, _ = Update(updated, MsgToolPrewarmFailed{Tool: "amp", Err: errTest("boom")})
	if updated.ToolWarmupCompleted != 3 {
		t.Fatalf("expected completed count to be 3 after failed message, got %d", updated.ToolWarmupCompleted)
	}
	if updated.ToolWarmupFailed != 1 {
		t.Fatalf("expected failed count to be 1 after failed message, got %d", updated.ToolWarmupFailed)
	}
}

func TestToolStartingUnhandledKeyIgnored(t *testing.T) {
	m := Model{Mode: ModeToolStarting}
	_, _, handled := UpdateKey(m, "a")
	if handled {
		t.Fatal("expected unbound key to be unhandled in tool-starting mode")
	}
}
