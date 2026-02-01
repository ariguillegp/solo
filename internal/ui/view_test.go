package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
	if !strings.Contains(view, "Delete Worktree") {
		t.Fatalf("expected delete prompt, got %q", view)
	}
	if !strings.Contains(view, "/repo/feature") {
		t.Fatalf("expected delete path, got %q", view)
	}
	if !strings.Contains(view, "enter") && !strings.Contains(view, "delete") {
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

func TestViewHelpLinePerMode(t *testing.T) {
	tests := []struct {
		name      string
		mode      core.Mode
		setup     func(m *Model)
		helpParts []string
	}{
		{
			name:      "loading",
			mode:      core.ModeLoading,
			helpParts: []string{"esc", "quit"},
		},
		{
			name: "browsing",
			mode: core.ModeBrowsing,
			setup: func(m *Model) {
				m.core.Query = "proj"
				m.input.SetValue("proj")
			},
			helpParts: []string{"navigate", "enter", "select", "ctrl+n", "create", "esc", "quit"},
		},
		{
			name: "worktree",
			mode: core.ModeWorktree,
			setup: func(m *Model) {
				m.core.WorktreeQuery = "feat"
				m.worktreeInput.SetValue("feat")
			},
			helpParts: []string{"navigate", "enter", "select", "ctrl+n", "create", "ctrl+d", "delete", "esc", "back"},
		},
		{
			name: "worktree delete confirm",
			mode: core.ModeWorktreeDeleteConfirm,
			setup: func(m *Model) {
				m.core.WorktreeDeletePath = "/repo/feature"
			},
			helpParts: []string{"enter", "delete", "esc", "cancel"},
		},
		{
			name: "tool",
			mode: core.ModeTool,
			setup: func(m *Model) {
				m.core.ToolQuery = "amp"
				m.toolInput.SetValue("amp")
			},
			helpParts: []string{"navigate", "enter", "open", "esc", "back"},
		},
		{
			name: "error",
			mode: core.ModeError,
			setup: func(m *Model) {
				m.core.Err = errTest("boom")
			},
			helpParts: []string{"esc", "quit"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := newTestModel()
			m.core.Mode = test.mode
			if test.setup != nil {
				test.setup(&m)
			}

			view := stripANSI(m.View())
			for _, part := range test.helpParts {
				if !strings.Contains(view, part) {
					t.Fatalf("expected help line to contain %q, got %q", part, view)
				}
			}
		})
	}
}

func TestViewStepHeaders(t *testing.T) {
	tests := []struct {
		name   string
		mode   core.Mode
		header string
	}{
		{"browsing", core.ModeBrowsing, "Step 1: Select Project"},
		{"worktree", core.ModeWorktree, "Step 2: Select Worktree"},
		{"worktree delete", core.ModeWorktreeDeleteConfirm, "Delete Worktree"},
		{"tool", core.ModeTool, "Step 3: Select Tool"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := newTestModel()
			m.core.Mode = test.mode
			if test.mode == core.ModeWorktreeDeleteConfirm {
				m.core.WorktreeDeletePath = "/repo/feature"
			}

			view := stripANSI(m.View())
			if !strings.Contains(view, test.header) {
				t.Fatalf("expected header %q, got %q", test.header, view)
			}
		})
	}
}

func TestViewScrollIndicators(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeBrowsing
	m.core.Query = "proj"
	m.input.SetValue("proj")

	var dirs []core.DirEntry
	for i := 1; i <= 10; i++ {
		dirs = append(dirs, core.DirEntry{Path: fmt.Sprintf("/proj%d", i), Name: fmt.Sprintf("proj%d", i)})
	}
	m.core.Filtered = dirs

	t.Run("top shows more below", func(t *testing.T) {
		m.core.SelectedIdx = 0
		view := stripANSI(m.View())
		if !strings.Contains(view, "more below") {
			t.Fatalf("expected 'more below' indicator at top, got %q", view)
		}
		if strings.Contains(view, "more above") {
			t.Fatalf("did not expect 'more above' indicator at top, got %q", view)
		}
	})

	t.Run("middle shows both indicators", func(t *testing.T) {
		m.core.SelectedIdx = 5
		view := stripANSI(m.View())
		if !strings.Contains(view, "more above") {
			t.Fatalf("expected 'more above' indicator in middle, got %q", view)
		}
		if !strings.Contains(view, "more below") {
			t.Fatalf("expected 'more below' indicator in middle, got %q", view)
		}
	})

	t.Run("bottom shows more above", func(t *testing.T) {
		m.core.SelectedIdx = 9
		view := stripANSI(m.View())
		if !strings.Contains(view, "more above") {
			t.Fatalf("expected 'more above' indicator at bottom, got %q", view)
		}
		if strings.Contains(view, "more below") {
			t.Fatalf("did not expect 'more below' indicator at bottom, got %q", view)
		}
	})
}

func TestViewNoScrollIndicatorsForShortList(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeBrowsing
	m.core.Query = "proj"
	m.input.SetValue("proj")
	m.core.Filtered = []core.DirEntry{
		{Path: "/proj1", Name: "proj1"},
		{Path: "/proj2", Name: "proj2"},
	}
	m.core.SelectedIdx = 0

	view := stripANSI(m.View())
	if strings.Contains(view, "more above") || strings.Contains(view, "more below") {
		t.Fatalf("did not expect scroll indicators for short list, got %q", view)
	}
}

func TestThemePickerEscRestoresTheme(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeBrowsing
	m.themeIdx = 0
	m.styles = NewStyles(m.themes[0])

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	m = model.(Model)
	if !m.showThemePicker {
		t.Fatalf("expected theme picker to open")
	}

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(Model)
	if m.themeIdx != 1 {
		t.Fatalf("expected theme idx to move down, got %d", m.themeIdx)
	}

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = model.(Model)
	if m.showThemePicker {
		t.Fatalf("expected theme picker to close")
	}
	if m.themeIdx != 0 {
		t.Fatalf("expected theme idx to restore, got %d", m.themeIdx)
	}
}

type errTest string

func (e errTest) Error() string {
	return string(e)
}
