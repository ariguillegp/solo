package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/ariguillegp/rivet/internal/core"
	"github.com/ariguillegp/rivet/internal/ui/listmodel"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	defaultListSuggestions = 5
	minListSuggestions     = 5
	maxListSuggestions     = 12
	listReservedLines      = 10
)

type suggestionItem struct {
	primary     string
	detail      string
	actionLabel string
}

func (i suggestionItem) FilterValue() string {
	return strings.TrimSpace(i.primary + " " + i.detail + " " + i.actionLabel)
}

type suggestionDelegate struct{ styles Styles }

func (d suggestionDelegate) Height() int                                  { return 1 }
func (d suggestionDelegate) Spacing() int                                 { return 0 }
func (d suggestionDelegate) Update(_ tea.Msg, _ *listmodel.Model) tea.Cmd { return nil }
func (d suggestionDelegate) Render(w io.Writer, m listmodel.Model, index int, item listmodel.Item) {
	row, ok := item.(suggestionItem)
	if !ok {
		return
	}
	selected := index == m.Index()
	prefix := "  "
	if selected {
		prefix = "> "
	}

	if row.actionLabel != "" {
		actionStyle := d.styles.Action
		valueStyle := d.styles.Suggestion
		pathStyle := d.styles.Path
		if selected {
			actionStyle = d.styles.SelectedAction
			valueStyle = d.styles.SelectedSuggestion
			pathStyle = d.styles.SelectedPath
		}
		parts := []string{actionStyle.Render(row.actionLabel), valueStyle.Render(row.primary)}
		if row.detail != "" {
			parts = append(parts, pathStyle.Render("- "+row.detail))
		}
		_, _ = fmt.Fprint(w, prefix+strings.Join(parts, " "))
		return
	}

	primaryStyle := d.styles.Suggestion
	pathStyle := d.styles.Path
	if selected {
		primaryStyle = d.styles.SelectedSuggestion
		pathStyle = d.styles.SelectedPath
	}
	parts := []string{primaryStyle.Render(row.primary)}
	if row.detail != "" {
		parts = append(parts, pathStyle.Render("- "+row.detail))
	}
	_, _ = fmt.Fprint(w, prefix+strings.Join(parts, " "))
}

func newSuggestionList(styles Styles) listmodel.Model {
	l := listmodel.New([]listmodel.Item{}, suggestionDelegate{styles: styles}, 0, defaultListSuggestions)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	l.KeyMap.CursorUp = key.NewBinding(key.WithKeys("up", "ctrl+k"))
	l.KeyMap.CursorDown = key.NewBinding(key.WithKeys("down", "ctrl+j"))
	return l
}

func (m Model) listLimit() int {
	if m.height <= 0 {
		return defaultListSuggestions
	}
	available := m.height - listReservedLines
	if available < minListSuggestions {
		return minListSuggestions
	}
	if available > maxListSuggestions {
		return maxListSuggestions
	}
	return available
}

func listHeight(limit, total int) int {
	if total <= 0 {
		return 0
	}
	if limit <= 0 {
		return total
	}
	if total < limit {
		return total
	}
	return limit
}

func visibleListWindow(total, selectedIdx, maxItems int) (start, end int) {
	if total <= 0 {
		return 0, 0
	}
	if maxItems <= 0 || maxItems > total {
		maxItems = total
	}
	if selectedIdx < 0 {
		selectedIdx = 0
	}
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
			start = end - maxItems
			if start < 0 {
				start = 0
			}
		}
	}
	return start, end
}

func (m Model) renderCount(l listmodel.Model) string {
	total := len(l.Items())
	if total == 0 {
		return ""
	}
	start, end := visibleListWindow(total, l.Index(), l.Height())
	if end == 0 {
		return ""
	}
	line := fmt.Sprintf("Showing %d-%d of %d", start+1, end, total)
	return "\n" + m.styles.Count.Render(line)
}

func toItems(rows []suggestionItem) []listmodel.Item {
	items := make([]listmodel.Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, row)
	}
	return items
}

func (m *Model) applyListStyles() {
	lists := []*listmodel.Model{&m.projectList, &m.worktreeList, &m.toolList, &m.sessionList, &m.themeList}
	for _, l := range lists {
		l.SetDelegate(suggestionDelegate{styles: m.styles})
		l.SetHeight(listHeight(m.listLimit(), len(l.Items())))
	}
}

func (m *Model) syncProjectList() {
	rows := make([]suggestionItem, 0, len(m.core.Filtered)+1)
	for _, dir := range m.core.Filtered {
		rows = append(rows, suggestionItem{primary: dir.Name, detail: m.displayPath(dir.Path)})
	}
	if createPath, ok := m.core.CreateProjectPath(); ok {
		rows = append(rows, suggestionItem{primary: m.displayPath(createPath), actionLabel: "create"})
	}
	m.projectList.SetItems(toItems(rows))
	m.projectList.SetHeight(listHeight(m.listLimit(), len(rows)))
	m.projectList.Select(m.core.SelectedIdx)
}

func (m *Model) syncWorktreeList() {
	rows := make([]suggestionItem, 0, len(m.core.FilteredWT)+1)
	for _, wt := range m.core.FilteredWT {
		rows = append(rows, suggestionItem{primary: m.worktreeDisplayLabel(wt)})
	}
	if name, ok := m.core.CreateWorktreeName(); ok {
		rows = append(rows, suggestionItem{primary: name, actionLabel: "create"})
	}
	m.worktreeList.SetItems(toItems(rows))
	m.worktreeList.SetHeight(listHeight(m.listLimit(), len(rows)))
	m.worktreeList.Select(m.core.WorktreeIdx)
}

func (m *Model) syncToolList() {
	rows := make([]suggestionItem, 0, len(m.core.FilteredTools))
	for _, tool := range m.core.FilteredTools {
		rows = append(rows, suggestionItem{primary: tool})
	}
	m.toolList.SetItems(toItems(rows))
	m.toolList.SetHeight(listHeight(m.listLimit(), len(rows)))
	m.toolList.Select(m.core.ToolIdx)
}

func (m *Model) syncSessionList() {
	rows := make([]suggestionItem, 0, len(m.core.FilteredSessions))
	for _, session := range m.core.FilteredSessions {
		label := core.SessionDisplayLabel(session)
		if label == "" {
			label = m.displayPath(session.DirPath)
		}
		rows = append(rows, suggestionItem{primary: label})
	}
	m.sessionList.SetItems(toItems(rows))
	m.sessionList.SetHeight(listHeight(m.listLimit(), len(rows)))
	m.sessionList.Select(m.core.SessionIdx)
}

func (m *Model) syncThemeList() {
	rows := make([]suggestionItem, 0, len(m.filteredThemes))
	for _, theme := range m.filteredThemes {
		rows = append(rows, suggestionItem{primary: theme.Name})
	}
	m.themeList.SetItems(toItems(rows))
	m.themeList.SetHeight(listHeight(m.listLimit(), len(rows)))
	idx := indexOfThemeByName(m.filteredThemes, m.themes[m.activeThemeIdx].Name)
	if idx < 0 {
		idx = 0
	}
	m.themeList.Select(idx)
}

func (m *Model) syncLists() {
	m.applyListStyles()
	m.syncProjectList()
	m.syncWorktreeList()
	m.syncToolList()
	m.syncSessionList()
	m.syncThemeList()
}
