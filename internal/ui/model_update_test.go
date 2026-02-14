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
