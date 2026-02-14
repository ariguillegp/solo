package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"

	"github.com/ariguillegp/rivet/internal/core"
)

func newTestModel() Model {
	m := New(nil, nil, nil)
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
	m.height = 25
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
	if !strings.Contains(view, "two - /two") {
		t.Fatalf("expected selected project suggestion, got %q", view)
	}
}

func TestViewBrowsingUsesTildeForHome(t *testing.T) {
	m := newTestModel()
	m.height = 25
	m.core.Mode = core.ModeBrowsing
	m.core.Query = "proj"
	m.input.SetValue("proj")
	m.core.Filtered = []core.DirEntry{
		{Path: "/home/demo/Projects/rivet", Name: "rivet"},
	}
	m.core.SelectedIdx = 0
	m.homeDir = "/home/demo"

	view := stripANSI(m.View())
	if !strings.Contains(view, "~/Projects/rivet") {
		t.Fatalf("expected home path to use tilde, got %q", view)
	}
	if strings.Contains(view, "/home/demo/Projects/rivet") {
		t.Fatalf("expected home path to be replaced, got %q", view)
	}
}

func TestViewBrowsingCreateNew(t *testing.T) {
	m := newTestModel()
	m.height = 25
	m.core.Mode = core.ModeBrowsing
	m.core.Query = "newproj"
	m.input.SetValue("newproj")
	m.core.RootPaths = []string{"/tmp"}
	m.core.Filtered = nil
	m.core.SelectedIdx = 0

	view := stripANSI(m.View())
	if !strings.Contains(view, "create  /tmp/newproj") {
		t.Fatalf("expected create new hint, got %q", view)
	}
}

func TestViewWorktreeSuggestionAndNav(t *testing.T) {
	m := newTestModel()
	m.height = 25
	m.core.Mode = core.ModeWorktree
	m.core.WorktreeQuery = "feat"
	m.worktreeInput.SetValue("feat")
	m.core.FilteredWT = []core.Worktree{
		{Path: "/repo/main", Name: "main", Branch: "main"},
		{Path: "/repo/feat", Name: "feat", Branch: "feat"},
	}
	m.core.WorktreeIdx = 1

	view := stripANSI(m.View())
	if !strings.Contains(view, "Select workspace or create new branch") {
		t.Fatalf("expected workspace prompt, got %q", view)
	}
	if !strings.Contains(view, "feat") {
		t.Fatalf("expected selected worktree suggestion, got %q", view)
	}
}

func TestViewWorktreeCreateNew(t *testing.T) {
	m := newTestModel()
	m.height = 25
	m.core.Mode = core.ModeWorktree
	m.core.WorktreeQuery = "feature-x"
	m.worktreeInput.SetValue("feature-x")
	m.core.FilteredWT = nil
	m.core.WorktreeIdx = 0
	m.core.WorktreeWarning = "worktree already exists for branch feature-x"

	view := stripANSI(m.View())
	if !strings.Contains(view, "create") {
		t.Fatalf("expected create new workspace hint, got %q", view)
	}
	if !strings.Contains(view, "feature-x") {
		t.Fatalf("expected create new workspace hint, got %q", view)
	}
	if !strings.Contains(view, "worktree already exists") {
		t.Fatalf("expected worktree warning to be shown, got %q", view)
	}
}

func TestViewWorktreeDeleteConfirm(t *testing.T) {
	m := newTestModel()
	m.height = 25
	m.core.Mode = core.ModeWorktreeDeleteConfirm
	m.core.WorktreeDeletePath = "/repo/feature"
	m.core.SelectedWorktreePath = "/repo/feature"
	m.core.Worktrees = []core.Worktree{{Path: "/repo/feature", Branch: "feature"}}

	view := stripANSI(m.View())
	if !strings.Contains(view, "Delete Workspace") {
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
	m.height = 25
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
	m.height = 25
	m.core.Mode = core.ModeTool
	m.core.ToolQuery = "missing"
	m.toolInput.SetValue("missing")
	m.core.FilteredTools = nil
	m.core.ToolIdx = 0

	view := stripANSI(m.View())
	if !strings.Contains(view, "No matches") {
		t.Fatalf("expected empty state, got %q", view)
	}
}

func TestViewEmptyState(t *testing.T) {
	m := newTestModel()
	m.height = 25
	m.core.Mode = core.ModeBrowsing
	m.core.Query = "missing"
	m.input.SetValue("missing")
	m.core.Filtered = nil
	m.core.SelectedIdx = 0

	view := stripANSI(m.View())
	if !strings.Contains(view, "No matches") {
		t.Fatalf("expected empty state to mention no matches, got %q", view)
	}
}

func TestViewToolStarting(t *testing.T) {
	m := newTestModel()
	m.height = 25
	m.core.Mode = core.ModeToolStarting
	m.core.PendingSpec = &core.SessionSpec{Tool: "opencode"}
	m.core.ToolWarmupTotal = 4
	m.core.ToolWarmupCompleted = 3
	m.core.ToolWarmupFailed = 1

	view := stripANSI(m.View())
	if !strings.Contains(view, "Starting opencode") {
		t.Fatalf("expected tool starting view, got %q", view)
	}
	if !strings.Contains(view, "%") {
		t.Fatalf("expected progress bar to include percentage text, got %q", view)
	}
}

func TestViewError(t *testing.T) {
	m := newTestModel()
	m.height = 25
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
			helpParts: []string{"enter", "select", "ctrl+s", "sessions", "?", "help", "esc", "quit"},
		},
		{
			name: "worktree",
			mode: core.ModeWorktree,
			setup: func(m *Model) {
				m.core.WorktreeQuery = "feat"
				m.worktreeInput.SetValue("feat")
			},
			helpParts: []string{"enter", "select", "ctrl+d", "delete", "ctrl+s", "sessions", "?", "help", "esc", "back"},
		},
		{
			name: "worktree delete confirm",
			mode: core.ModeWorktreeDeleteConfirm,
			setup: func(m *Model) {
				m.core.WorktreeDeletePath = "/repo/feature"
				m.core.SelectedWorktreePath = "/repo/feature"
				m.core.Worktrees = []core.Worktree{{Path: "/repo/feature", Branch: "feature"}}
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
			helpParts: []string{"enter", "open", "ctrl+s", "sessions", "?", "help", "esc", "back"},
		},
		{
			name:      "tool starting",
			mode:      core.ModeToolStarting,
			helpParts: []string{"esc", "cancel", "ctrl+c", "quit"},
		},
		{
			name: "sessions",
			mode: core.ModeSessions,
			setup: func(m *Model) {
				sessions := []core.SessionInfo{{Name: "session"}}
				m.core.Sessions = sessions
				m.core.FilteredSessions = sessions
			},
			helpParts: []string{"enter", "attach", "?", "help", "esc", "back"},
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
			m.height = 25
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

func TestViewSessionsEmptyHelpOmitsAttach(t *testing.T) {
	m := newTestModel()
	m.height = 25
	m.core.Mode = core.ModeSessions
	m.core.Sessions = nil
	m.core.FilteredSessions = nil

	view := stripANSI(m.View())
	if strings.Contains(view, "attach") {
		t.Fatalf("expected empty sessions help to omit attach action, got %q", view)
	}
	for _, part := range []string{"?", "help", "esc", "back"} {
		if !strings.Contains(view, part) {
			t.Fatalf("expected empty sessions help to contain %q, got %q", part, view)
		}
	}
}

func TestViewThemePicker(t *testing.T) {
	m := newTestModel()
	m.height = 25
	m.showThemePicker = true
	m.filteredThemes = []Theme{{Name: "Gruvbox"}, {Name: "Dracula"}}
	m.syncThemeList()
	m.themeList.Select(1)
	m.themeInput.SetValue("dra")

	view := stripANSI(m.View())
	if !strings.Contains(view, "Theme Picker") {
		t.Fatalf("expected theme picker header, got %q", view)
	}
	if !strings.Contains(view, "Filter themes") {
		t.Fatalf("expected theme picker prompt, got %q", view)
	}
	if !strings.Contains(view, "Dracula") {
		t.Fatalf("expected theme list to include Dracula, got %q", view)
	}
}

func TestViewHelpModalPerMode(t *testing.T) {
	tests := []struct {
		name      string
		mode      core.Mode
		wantParts []string
	}{
		{
			name:      "browsing help",
			mode:      core.ModeBrowsing,
			wantParts: []string{"enter", "select", "ctrl+s", "sessions", "ctrl+d", "delete", "ctrl+t", "theme"},
		},
		{
			name:      "delete confirm help",
			mode:      core.ModeWorktreeDeleteConfirm,
			wantParts: []string{"enter", "delete", "esc", "cancel", "ctrl+c", "quit"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := newTestModel()
			m.height = 25
			m.core.Mode = test.mode
			m.showHelp = true

			view := stripANSI(m.View())
			for _, part := range test.wantParts {
				if !strings.Contains(view, part) {
					t.Fatalf("expected help modal to contain %q, got %q", part, view)
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
		{"worktree", core.ModeWorktree, "Step 2: Select Workspace"},
		{"worktree delete", core.ModeWorktreeDeleteConfirm, "Delete Workspace"},
		{"tool", core.ModeTool, "Step 3: Select Tool"},
		{"sessions", core.ModeSessions, "Active tmux sessions"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := newTestModel()
			m.height = 25
			m.core.Mode = test.mode
			if test.mode == core.ModeWorktreeDeleteConfirm {
				m.core.WorktreeDeletePath = "/repo/feature"
				m.core.SelectedWorktreePath = "/repo/feature"
				m.core.Worktrees = []core.Worktree{{Path: "/repo/feature", Branch: "feature"}}
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
		if !strings.Contains(view, "proj10") {
			t.Fatalf("expected selected item to be visible in rendered list, got %q", view)
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

type errTest string

func (e errTest) Error() string {
	return string(e)
}

func TestToolStartingProgressUsesWarmupDelayFraction(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeToolStarting
	toolReadyDelay := core.ToolWarmupDelay("")
	m.toolStartingAt = time.Now().Add(-(toolReadyDelay / 2))
	m.toolStartingDuration = toolReadyDelay

	progress := m.toolStartingProgress()
	if progress <= 0.35 || progress >= 0.65 {
		t.Fatalf("expected in-progress value around halfway, got %f", progress)
	}
}

func TestToolStartingProgressStaysBelowOneForExistingSession(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeToolStarting
	toolReadyDelay := core.ToolWarmupDelay("")
	m.toolStartingAt = time.Now().Add(-toolReadyDelay)
	m.toolStartingDuration = toolReadyDelay

	progress := m.toolStartingProgress()
	if progress != maxDisplayedToolStartingProgress {
		t.Fatalf("expected progress to stay below complete while startup is in-flight, got %f", progress)
	}
}

func TestToolStartingProgressStaysBelowOneWhileWaitingForOpen(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeToolStarting
	toolReadyDelay := core.ToolWarmupDelay("")
	m.toolStartingAt = time.Now().Add(-toolReadyDelay)
	m.toolStartingDuration = toolReadyDelay

	progress := m.toolStartingProgress()
	if progress != maxDisplayedToolStartingProgress {
		t.Fatalf("expected progress to stay below complete while opening session, got %f", progress)
	}
}

func TestViewSessionsUsesTableOnWideTerminal(t *testing.T) {
	m := newTestModel()
	m.height = 25
	m.width = 140
	m.core.Mode = core.ModeSessions
	m.core.Sessions = []core.SessionInfo{{
		Name:       "alpha",
		DirPath:    "/home/demo/Projects/rivet/rbac-sentinel",
		Project:    "/home/demo/Projects/rivet",
		Branch:     "rbac-sentinel",
		Tool:       "codex",
		LastActive: time.Unix(1735689600, 0),
	}}
	m.homeDir = "/home/demo"
	m.core.FilteredSessions = m.core.Sessions
	m.syncSessionList()

	view := stripANSI(m.View())
	for _, part := range []string{"Project", "Branch", "Tool", "Last active", "~/Projects/rivet", "rbac-sentinel"} {
		if !strings.Contains(view, part) {
			t.Fatalf("expected wide sessions view to contain %q, got %q", part, view)
		}
	}
}

func TestViewSessionsUsesCompactListOnNarrowTerminal(t *testing.T) {
	m := newTestModel()
	m.height = 25
	m.width = 80
	m.core.Mode = core.ModeSessions
	m.core.Sessions = []core.SessionInfo{{Name: "alpha", DirPath: "/repo/project/feature-a", Tool: "codex"}}
	m.core.FilteredSessions = m.core.Sessions
	m.syncSessionList()

	view := stripANSI(m.View())
	if strings.Contains(view, "Branch") {
		t.Fatalf("expected narrow sessions view to omit table headers, got %q", view)
	}
	if !strings.Contains(view, "alpha") {
		t.Fatalf("expected narrow sessions view to include session label, got %q", view)
	}
}
