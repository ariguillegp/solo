package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ariguillegp/rivet/internal/core"
)

func TestUpdateKeyNavigationMovesOnceInBrowsing(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeBrowsing
	m.core.Filtered = []core.DirEntry{
		{Path: "/one", Name: "one"},
		{Path: "/two", Name: "two"},
		{Path: "/three", Name: "three"},
	}
	m.core.SelectedIdx = 0
	m.syncLists()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	next := updated.(Model)
	if next.core.SelectedIdx != 1 {
		t.Fatalf("expected single-step move to index 1, got %d", next.core.SelectedIdx)
	}

	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyDown})
	next = updated.(Model)
	if next.core.SelectedIdx != 2 {
		t.Fatalf("expected single-step move to index 2, got %d", next.core.SelectedIdx)
	}
}

func runUpdate(model Model, msg tea.Msg) Model {
	updated, _ := model.Update(msg)
	return updated.(Model)
}

func TestPaletteOpenSessionsMatchesShortcut(t *testing.T) {
	direct := newTestModel()
	direct.core.Mode = core.ModeBrowsing
	direct = runUpdate(direct, tea.KeyMsg{Type: tea.KeyCtrlS})

	palette := newTestModel()
	palette.core.Mode = core.ModeBrowsing
	palette = runUpdate(palette, tea.KeyMsg{Type: tea.KeyCtrlP})
	palette = runUpdate(palette, tea.KeyMsg{Type: tea.KeyEnter})

	if direct.core.Mode != palette.core.Mode {
		t.Fatalf("expected same mode after open sessions, direct=%v palette=%v", direct.core.Mode, palette.core.Mode)
	}
}

func TestPaletteDeleteCurrentMatchesShortcut(t *testing.T) {
	direct := newTestModel()
	direct.core.Mode = core.ModeBrowsing
	direct.core.Filtered = []core.DirEntry{{Path: "/repo", Name: "repo"}}
	direct.core.SelectedIdx = 0
	direct.syncLists()
	direct = runUpdate(direct, tea.KeyMsg{Type: tea.KeyCtrlD})

	palette := newTestModel()
	palette.core.Mode = core.ModeBrowsing
	palette.core.Filtered = []core.DirEntry{{Path: "/repo", Name: "repo"}}
	palette.core.SelectedIdx = 0
	palette.syncLists()
	palette = runUpdate(palette, tea.KeyMsg{Type: tea.KeyCtrlP})
	palette = runUpdate(palette, tea.KeyMsg{Type: tea.KeyDown})
	palette = runUpdate(palette, tea.KeyMsg{Type: tea.KeyDown})
	palette = runUpdate(palette, tea.KeyMsg{Type: tea.KeyEnter})

	if direct.core.Mode != palette.core.Mode {
		t.Fatalf("expected same mode after delete action, direct=%v palette=%v", direct.core.Mode, palette.core.Mode)
	}
	if direct.core.ProjectDeletePath != palette.core.ProjectDeletePath {
		t.Fatalf("expected same delete target, direct=%q palette=%q", direct.core.ProjectDeletePath, palette.core.ProjectDeletePath)
	}
}

func TestPaletteBackMatchesShortcut(t *testing.T) {
	direct := newTestModel()
	direct.core.Mode = core.ModeWorktree
	direct = runUpdate(direct, tea.KeyMsg{Type: tea.KeyEsc})

	palette := newTestModel()
	palette.core.Mode = core.ModeWorktree
	palette = runUpdate(palette, tea.KeyMsg{Type: tea.KeyCtrlP})
	palette = runUpdate(palette, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("back")})
	palette = runUpdate(palette, tea.KeyMsg{Type: tea.KeyEnter})

	if direct.core.Mode != palette.core.Mode {
		t.Fatalf("expected same mode after back action, direct=%v palette=%v", direct.core.Mode, palette.core.Mode)
	}
}
