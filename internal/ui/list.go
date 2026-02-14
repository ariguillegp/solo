package ui

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/ariguillegp/rivet/internal/core"
	"github.com/ariguillegp/rivet/internal/ui/listmodel"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
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

const compactSessionMinWidth = 95

func newSessionTable(styles Styles) table.Model {
	columns := []table.Column{
		{Title: "Project", Width: 40},
		{Title: "Branch", Width: 22},
		{Title: "Tool", Width: 12},
		{Title: "Last active", Width: 16},
	}
	t := table.New(table.WithColumns(columns), table.WithRows(nil), table.WithFocused(true), table.WithHeight(defaultListSuggestions))
	t.SetStyles(newSessionTableStyles(styles))
	t.KeyMap.LineUp = key.NewBinding(key.WithKeys("up", "ctrl+k"))
	t.KeyMap.LineDown = key.NewBinding(key.WithKeys("down", "ctrl+j"))
	return t
}

func newSessionTableStyles(styles Styles) table.Styles {
	ts := table.DefaultStyles()
	headerFg := styles.Body.GetForeground()
	cellFg := styles.Path.GetForeground()
	selectedFg := styles.SelectedSuggestion.GetForeground()

	ts.Header = ts.Header.Bold(true)
	if headerFg != nil {
		ts.Header = ts.Header.Foreground(headerFg)
	}
	if cellFg != nil {
		ts.Cell = ts.Cell.Foreground(cellFg)
	}

	ts.Selected = ts.Cell.Copy().UnsetBackground().Bold(true)
	if selectedFg != nil {
		ts.Selected = ts.Selected.Foreground(selectedFg)
	}
	return ts
}

func (m Model) sessionListIsCompact() bool {
	return m.width > 0 && m.width < compactSessionMinWidth
}

func sessionLastActiveLabel(lastActive time.Time) string {
	if lastActive.IsZero() {
		return "—"
	}
	return lastActive.Local().Format("2006-01-02 15:04")
}

func (m Model) sessionProjectLabel(session core.SessionInfo) string {
	project := strings.TrimSpace(session.Project)
	if project != "" {
		return project
	}
	if session.DirPath == "" {
		return ""
	}
	projectPath := filepath.Dir(session.DirPath)
	if projectPath == "." || projectPath == string(filepath.Separator) {
		return ""
	}
	return filepath.Base(projectPath)
}

func (m Model) sessionBranchLabel(session core.SessionInfo) string {
	if branch := strings.TrimSpace(session.Branch); branch != "" {
		return branch
	}
	name := core.SessionWorktreeName(session.DirPath)
	if name != "" {
		return name
	}
	return filepath.Base(session.DirPath)
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
	m.sessionTable.SetStyles(newSessionTableStyles(m.styles))
	m.sessionTable.SetHeight(listHeight(m.listLimit(), len(m.sessionTable.Rows())))
	if m.width > 0 {
		m.sessionTable.SetWidth(max(0, m.width-8))
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
	compactRows := make([]suggestionItem, 0, len(m.core.FilteredSessions))
	tableRows := make([]table.Row, 0, len(m.core.FilteredSessions))
	for _, session := range m.core.FilteredSessions {
		label := core.SessionDisplayLabel(session)
		if label == "" {
			label = m.displayPath(session.DirPath)
		}
		compactRows = append(compactRows, suggestionItem{primary: label})
		tableRows = append(tableRows, table.Row{
			m.sessionProjectLabel(session),
			m.sessionBranchLabel(session),
			session.Tool,
			sessionLastActiveLabel(session.LastActive),
		})
	}
	m.sessionList.SetItems(toItems(compactRows))
	m.sessionList.SetHeight(listHeight(m.listLimit(), len(compactRows)))
	m.sessionList.Select(m.core.SessionIdx)
	m.sessionTable.SetRows(tableRows)
	m.sessionTable.SetHeight(listHeight(m.listLimit(), len(tableRows)))
	m.sessionTable.SetCursor(m.core.SessionIdx)
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
