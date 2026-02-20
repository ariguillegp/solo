package listmodel

import (
	"bytes"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Item interface {
	FilterValue() string
}

type ItemDelegate interface {
	Render(w io.Writer, m Model, index int, item Item)
	Height() int
	Spacing() int
	Update(msg tea.Msg, m *Model) tea.Cmd
}

type KeyMap struct {
	CursorUp   key.Binding
	CursorDown key.Binding
}

type Model struct {
	items    []Item
	delegate ItemDelegate
	cursor   int
	height   int
	KeyMap   KeyMap
}

func New(items []Item, delegate ItemDelegate, _, height int) Model {
	m := Model{items: items, delegate: delegate, height: height}
	m.KeyMap = KeyMap{
		CursorUp:   key.NewBinding(key.WithKeys("up", "k")),
		CursorDown: key.NewBinding(key.WithKeys("down", "j")),
	}
	return m
}

func (m *Model) SetShowTitle(bool)      {}
func (m *Model) SetShowStatusBar(bool)  {}
func (m *Model) SetShowPagination(bool) {}
func (m *Model) SetShowHelp(bool)       {}
func (m *Model) SetFilteringEnabled(bool) {
}
func (m *Model) DisableQuitKeybindings() {}

func (m *Model) SetDelegate(d ItemDelegate) { m.delegate = d }
func (m *Model) SetHeight(h int) {
	if h < 0 {
		h = 0
	}
	m.height = h
}
func (m Model) Height() int { return m.height }

func (m *Model) SetItems(items []Item) {
	m.items = items
	if len(m.items) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
}
func (m Model) Items() []Item { return m.items }

func (m *Model) Select(idx int) {
	if len(m.items) == 0 {
		m.cursor = 0
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(m.items) {
		idx = len(m.items) - 1
	}
	m.cursor = idx
}

func (m Model) Index() int { return m.cursor }

func (m Model) visibleWindow() (start, end int) {
	total := len(m.items)
	if total == 0 {
		return 0, 0
	}
	maxItems := m.height
	if maxItems <= 0 || maxItems > total {
		maxItems = total
	}
	selectedIdx := max(m.cursor, 0)
	if selectedIdx >= total {
		selectedIdx = total - 1
	}
	start = 0
	if selectedIdx >= maxItems {
		start = selectedIdx - maxItems + 1
	}
	end = start + maxItems
	if end > total {
		end = total
		if end-start < maxItems && start > 0 {
			start = max(end-maxItems, 0)
		}
	}
	return start, end
}

func (m Model) View() string {
	if len(m.items) == 0 || m.delegate == nil {
		return ""
	}
	start, end := m.visibleWindow()
	var b strings.Builder
	if start > 0 {
		b.WriteString("  ▲ more above\n")
	}
	for i := start; i < end; i++ {
		if i > start {
			b.WriteString("\n")
		}
		var row bytes.Buffer
		m.delegate.Render(&row, m, i, m.items[i])
		b.WriteString(row.String())
	}
	if end < len(m.items) {
		b.WriteString("\n\n  ▼ more below")
	}
	return b.String()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.delegate != nil {
		if cmd := m.delegate.Update(msg, &m); cmd != nil {
			return m, cmd
		}
	}
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(keyMsg, m.KeyMap.CursorUp):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(keyMsg, m.KeyMap.CursorDown):
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	}
	return m, nil
}
