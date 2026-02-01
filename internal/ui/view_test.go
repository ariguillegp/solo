package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/ariguillegp/solo/internal/core"
)

func newTestModel() Model {
	m := New(nil, nil)
	m.width = 0
	m.height = 0
	return m
}

func stripANSI(input string) string {
	return ansi.Strip(input)
}

func TestViewLoading(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeLoading

	view := stripANSI(m.View())
	if !strings.Contains(view, "Scanning...") {
		t.Fatalf("expected loading view to contain scanning message, got %q", view)
	}
}

func TestViewBrowsingSuggestionAndNav(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeBrowsing
	m.core.Query = "proj"
	m.input.SetValue("proj")
	m.core.Filtered = []core.DirEntry{
		{Path: "/one", Name: "one"},
		{Path: "/two", Name: "two"},
	}
	m.core.SelectedIdx = 1

	view := stripANSI(m.View())
	if !strings.Contains(view, "Enter the project directory") {
		t.Fatalf("expected browsing prompt, got %q", view)
	}
	if !strings.Contains(view, "/two") {
		t.Fatalf("expected selected project suggestion, got %q", view)
	}
	if !strings.Contains(view, "[2/2]") {
		t.Fatalf("expected navigation hint, got %q", view)
	}
}

func TestViewBrowsingCreateNew(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeBrowsing
	m.core.Query = "newproj"
	m.input.SetValue("newproj")
	m.core.Filtered = nil
	m.core.SelectedIdx = 0

	view := stripANSI(m.View())
	if !strings.Contains(view, "(create new)") {
		t.Fatalf("expected create new hint, got %q", view)
	}
}

func TestViewWorktreeSuggestionAndNav(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeWorktree
	m.core.WorktreeQuery = "feat"
	m.worktreeInput.SetValue("feat")
	m.core.FilteredWT = []core.Worktree{
		{Path: "/repo/main", Name: "main", Branch: "main"},
		{Path: "/repo/feat", Name: "feat", Branch: "feat"},
	}
	m.core.WorktreeIdx = 1

	view := stripANSI(m.View())
	if !strings.Contains(view, "Select worktree or create new branch") {
		t.Fatalf("expected worktree prompt, got %q", view)
	}
	if !strings.Contains(view, "/repo/feat") {
		t.Fatalf("expected selected worktree suggestion, got %q", view)
	}
	if !strings.Contains(view, "[2/2]") {
		t.Fatalf("expected navigation hint, got %q", view)
	}
}

func TestViewWorktreeCreateNew(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeWorktree
	m.core.WorktreeQuery = "feature-x"
	m.worktreeInput.SetValue("feature-x")
	m.core.FilteredWT = nil
	m.core.WorktreeIdx = 0

	view := stripANSI(m.View())
	if !strings.Contains(view, "(create new: feature-x)") {
		t.Fatalf("expected create new worktree hint, got %q", view)
	}
}

func TestViewWorktreeDeleteConfirm(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeWorktreeDeleteConfirm
	m.core.WorktreeDeletePath = "/repo/feature"

	view := stripANSI(m.View())
	if !strings.Contains(view, "Delete worktree?") {
		t.Fatalf("expected delete prompt, got %q", view)
	}
	if !strings.Contains(view, "/repo/feature") {
		t.Fatalf("expected delete path, got %q", view)
	}
	if !strings.Contains(view, "enter to confirm") {
		t.Fatalf("expected confirmation hint, got %q", view)
	}
}

func TestViewToolSuggestionAndNav(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeTool
	m.core.ToolQuery = "amp"
	m.toolInput.SetValue("amp")
	m.core.FilteredTools = []string{"opencode", "amp"}
	m.core.ToolIdx = 1

	view := stripANSI(m.View())
	if !strings.Contains(view, "Select tool") {
		t.Fatalf("expected tool prompt, got %q", view)
	}
	if !strings.Contains(view, "amp") {
		t.Fatalf("expected selected tool suggestion, got %q", view)
	}
	if !strings.Contains(view, "[2/2]") {
		t.Fatalf("expected navigation hint, got %q", view)
	}
}

func TestViewToolNoCreateNew(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeTool
	m.core.ToolQuery = "missing"
	m.toolInput.SetValue("missing")
	m.core.FilteredTools = nil
	m.core.ToolIdx = 0

	view := stripANSI(m.View())
	if strings.Contains(view, "create new") {
		t.Fatalf("did not expect create new hint, got %q", view)
	}
}

func TestViewError(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeError
	m.core.Err = errTest("boom")

	view := stripANSI(m.View())
	if !strings.Contains(view, "Error: boom") {
		t.Fatalf("expected error view, got %q", view)
	}
}

type errTest string

func (e errTest) Error() string {
	return string(e)
}
