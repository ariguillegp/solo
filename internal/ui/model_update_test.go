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

func TestUpdateHelpCloseRestoresWorktreeFocus(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeWorktree
	m.showHelp = true
	m.blurInputs()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	next := updated.(Model)
	if next.showHelp {
		t.Fatalf("expected help to close")
	}
	if !next.worktreeInput.Focused() {
		t.Fatalf("expected worktree input to be focused after closing help")
	}
}

func TestUpdateViewportScrollHandledInHelp(t *testing.T) {
	m := newTestModel()
	m.showHelp = true
	m.width = 80
	m.height = 8
	m.updateViewportSize()
	m.viewport.SetContent("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\nline11\nline12\nline13\nline14\nline15")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	next := updated.(Model)
	if next.viewport.YOffset <= 0 {
		t.Fatalf("expected viewport to scroll down, got offset %d", next.viewport.YOffset)
	}
}

func TestUpdateViewportScrollHandledInDeleteConfirm(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeProjectDeleteConfirm
	m.width = 80
	m.height = 8
	m.updateViewportSize()
	m.viewport.SetContent("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\nline11\nline12\nline13\nline14\nline15")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	next := updated.(Model)
	if next.viewport.YOffset <= 0 {
		t.Fatalf("expected viewport to scroll down, got offset %d", next.viewport.YOffset)
	}
}

func TestUpdateTypingGInBrowsingUpdatesInput(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeBrowsing
	m.input.SetValue("")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	next := updated.(Model)
	if next.input.Value() != "g" {
		t.Fatalf("expected input value to include typed rune, got %q", next.input.Value())
	}
}

func TestUpdateViewportHomeEndHandledInHelp(t *testing.T) {
	m := newTestModel()
	m.showHelp = true
	m.width = 80
	m.height = 8
	m.updateViewportSize()
	m.viewport.SetContent(`line1
line2
line3
line4
line5
line6
line7
line8
line9
line10
line11
line12
line13
line14
line15`)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	next := updated.(Model)
	if next.viewport.YOffset <= 0 {
		t.Fatalf("expected end key to move viewport to bottom, got offset %d", next.viewport.YOffset)
	}

	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyHome})
	next = updated.(Model)
	if next.viewport.YOffset != 0 {
		t.Fatalf("expected home key to move viewport to top, got offset %d", next.viewport.YOffset)
	}
}

func TestUpdateEndMovesToLastProjectInBrowsing(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeBrowsing
	m.core.Filtered = []core.DirEntry{{Path: "/one", Name: "one"}, {Path: "/two", Name: "two"}, {Path: "/three", Name: "three"}}
	m.core.SelectedIdx = 0
	m.syncLists()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	next := updated.(Model)
	if next.core.SelectedIdx != 2 {
		t.Fatalf("expected end to move selection to last index 2, got %d", next.core.SelectedIdx)
	}
}

func TestUpdatePageDownMovesSelectionInBrowsing(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeBrowsing
	m.core.Filtered = []core.DirEntry{
		{Path: "/1", Name: "1"}, {Path: "/2", Name: "2"}, {Path: "/3", Name: "3"},
		{Path: "/4", Name: "4"}, {Path: "/5", Name: "5"}, {Path: "/6", Name: "6"},
	}
	m.core.SelectedIdx = 0
	m.syncLists()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	next := updated.(Model)
	if next.core.SelectedIdx <= 0 {
		t.Fatalf("expected pgdown to move selection forward, got %d", next.core.SelectedIdx)
	}
}
