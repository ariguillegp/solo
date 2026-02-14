package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ariguillegp/rivet/internal/core"
)

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Select   key.Binding
	Delete   key.Binding
	Sessions key.Binding
	Toggle   key.Binding
	Back     key.Binding
	Quit     key.Binding
	Theme    key.Binding
	Type     key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Up:       key.NewBinding(key.WithKeys("up", "ctrl+k"), key.WithHelp("↑/ctrl+k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "ctrl+j"), key.WithHelp("↓/ctrl+j", "down")),
		PageUp:   key.NewBinding(key.WithKeys("pgup", "pageup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("pgdown", "pgdn", "pagedown"), key.WithHelp("pgdn", "page down")),
		Top:      key.NewBinding(key.WithKeys("home"), key.WithHelp("home", "top")),
		Bottom:   key.NewBinding(key.WithKeys("end"), key.WithHelp("end", "bottom")),
		Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Delete:   key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "delete")),
		Sessions: key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "sessions")),
		Toggle:   key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Back:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Quit:     key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
		Theme:    key.NewBinding(key.WithKeys("ctrl+t"), key.WithHelp("ctrl+t", "theme")),
		Type:     key.NewBinding(key.WithKeys("type"), key.WithHelp("type", "filter")),
	}
}

func (k keyMap) actionForCore(msg tea.KeyMsg) (core.KeyAction, bool) {
	switch {
	case key.Matches(msg, k.Up):
		return core.KeyUp, true
	case key.Matches(msg, k.Down):
		return core.KeyDown, true
	case key.Matches(msg, k.PageUp):
		return core.KeyPageUp, true
	case key.Matches(msg, k.PageDown):
		return core.KeyPageDown, true
	case key.Matches(msg, k.Top):
		return core.KeyTop, true
	case key.Matches(msg, k.Bottom):
		return core.KeyBottom, true
	case key.Matches(msg, k.Select):
		return core.KeyEnter, true
	case key.Matches(msg, k.Delete):
		return core.KeyDelete, true
	case key.Matches(msg, k.Sessions):
		return core.KeySessions, true
	case key.Matches(msg, k.Back):
		return core.KeyBack, true
	case key.Matches(msg, k.Quit):
		return core.KeyQuit, true
	default:
		return "", false
	}
}

func (k keyMap) shortHelp(mode core.Mode) []key.Binding {
	switch mode {
	case core.ModeLoading:
		return []key.Binding{k.binding(k.Back, "quit")}
	case core.ModeBrowsing:
		return []key.Binding{k.Select, k.Delete, k.Sessions, k.Toggle, k.binding(k.Back, "quit")}
	case core.ModeWorktree:
		return []key.Binding{k.Select, k.Delete, k.Sessions, k.Toggle, k.Back}
	case core.ModeTool:
		return []key.Binding{k.binding(k.Select, "open"), k.Sessions, k.Toggle, k.Back}
	case core.ModeToolStarting:
		return []key.Binding{k.binding(k.Back, "cancel"), k.Quit}
	case core.ModeSessions:
		return []key.Binding{k.binding(k.Select, "attach"), k.Toggle, k.Back}
	default:
		return []key.Binding{k.binding(k.Back, "quit")}
	}
}

func (k keyMap) sessionsEmptyShortHelp() []key.Binding {
	return []key.Binding{k.Toggle, k.Back}
}

func (k keyMap) fullHelp(mode core.Mode) [][]key.Binding {
	common := [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown, k.Select},
		{k.Top, k.Bottom, k.Back, k.Quit, k.Toggle},
		{k.Type, k.Sessions, k.Delete, k.Theme},
	}
	switch mode {
	case core.ModeLoading, core.ModeError:
		return [][]key.Binding{{k.binding(k.Back, "quit"), k.Quit}, {k.Toggle}}
	case core.ModeProjectDeleteConfirm, core.ModeWorktreeDeleteConfirm:
		return [][]key.Binding{{k.binding(k.Select, "delete"), k.binding(k.Back, "cancel"), k.Quit}}
	case core.ModeToolStarting:
		return [][]key.Binding{{k.Back, k.Quit}}
	default:
		return common
	}
}

func (k keyMap) binding(b key.Binding, desc string) key.Binding {
	copy := b
	h := copy.Help()
	copy.SetHelp(h.Key, desc)
	return copy
}

func (m Model) shortHelpView() string {
	return m.help.ShortHelpView(m.keymap.shortHelp(m.core.Mode))
}

func (m Model) sessionsEmptyShortHelpView() string {
	return m.help.ShortHelpView(m.keymap.sessionsEmptyShortHelp())
}

func (m Model) fullHelpView() string {
	return m.help.FullHelpView(m.keymap.fullHelp(m.core.Mode))
}

func newHelpModel() help.Model {
	return help.New()
}
