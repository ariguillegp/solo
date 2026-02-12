package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ariguillegp/rivet/internal/core"
)

func (m Model) renderHelpLine(items []struct{ key, desc string }) string {
	var parts []string
	for _, item := range items {
		parts = append(parts, m.styles.Key.Render(item.key)+" "+m.styles.Help.Render(item.desc))
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, strings.Join(parts, "  "))
}

func (m Model) renderHelpRow(key, desc string) string {
	return m.styles.Key.Render(key) + " " + m.styles.Help.Render(desc)
}

const (
	defaultListSuggestions = 5
	minListSuggestions     = 5
	maxListSuggestions     = 12
	listReservedLines      = 10
)

type suggestionRow struct {
	primary     string
	detail      string
	actionLabel string
}

type suggestionWindow struct {
	start int
	end   int
	total int
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

func listWindow(total, selectedIdx, maxItems int) suggestionWindow {
	if total == 0 {
		return suggestionWindow{}
	}
	if maxItems <= 0 {
		maxItems = defaultListSuggestions
	}
	if maxItems > total {
		maxItems = total
	}
	if selectedIdx < 0 {
		selectedIdx = 0
	}
	if selectedIdx >= total {
		selectedIdx = total - 1
	}

	start := 0
	if selectedIdx >= maxItems {
		start = selectedIdx - maxItems + 1
	}
	end := start + maxItems
	if end > total {
		end = total
		if end-start < maxItems && start > 0 {
			start = end - maxItems
			if start < 0 {
				start = 0
			}
		}
	}

	return suggestionWindow{start: start, end: end, total: total}
}

func (m Model) renderSuggestionRow(row suggestionRow, selected bool) string {
	if row.actionLabel != "" {
		actionStyle := m.styles.Action
		valueStyle := m.styles.Suggestion
		pathStyle := m.styles.Path
		if selected {
			actionStyle = m.styles.SelectedAction
			valueStyle = m.styles.SelectedSuggestion
			pathStyle = m.styles.SelectedPath
		}

		parts := []string{actionStyle.Render(row.actionLabel), valueStyle.Render(row.primary)}
		if row.detail != "" {
			parts = append(parts, pathStyle.Render("- "+row.detail))
		}
		return strings.Join(parts, " ")
	}

	primaryStyle := m.styles.Suggestion
	pathStyle := m.styles.Path
	if selected {
		primaryStyle = m.styles.SelectedSuggestion
		pathStyle = m.styles.SelectedPath
	}
	parts := []string{primaryStyle.Render(row.primary)}
	if row.detail != "" {
		parts = append(parts, pathStyle.Render("- "+row.detail))
	}
	return strings.Join(parts, " ")
}

func (m Model) renderSuggestionList(rows []suggestionRow, selectedIdx, maxItems int) (string, suggestionWindow) {
	if len(rows) == 0 {
		return "", suggestionWindow{}
	}

	window := listWindow(len(rows), selectedIdx, maxItems)
	var out strings.Builder

	if window.start > 0 {
		out.WriteString(m.styles.ScrollIndicator.Render("  ▲ more above"))
		out.WriteString("\n")
	}

	for i := window.start; i < window.end; i++ {
		if i > window.start || window.start > 0 {
			out.WriteString("\n")
		}

		prefix := "  "
		selected := i == selectedIdx
		if selected {
			prefix = "> "
		}

		row := m.renderSuggestionRow(rows[i], selected)
		out.WriteString(prefix + row)
	}

	if window.end < len(rows) {
		out.WriteString("\n\n")
		out.WriteString(m.styles.ScrollIndicator.Render("  ▼ more below"))
	}

	return out.String(), window
}

func (m Model) View() string {
	if m.showHelp {
		return m.renderHelpModal()
	}
	if m.showThemePicker {
		return m.renderThemePicker()
	}

	var content string
	var helpLine string
	var header string
	var breadcrumb string

	switch m.core.Mode {
	case core.ModeLoading:
		content = m.spinner.View() + " Scanning..."
		helpLine = m.renderHelpLine([]struct{ key, desc string }{{"esc", "quit"}})

	case core.ModeBrowsing:
		header = m.styles.Title.Render("Step 1: Select Project")
		breadcrumb = m.renderBreadcrumb()
		prompt := m.styles.Prompt.Render("Enter the project directory:")
		input := prompt + " " + m.input.View()
		createPath, canCreate := m.core.CreateProjectPath()
		createLabel := ""
		createRow := suggestionRow{}
		if canCreate {
			createLabel = createPath
			createRow = suggestionRow{primary: m.displayPath(createPath), actionLabel: "create"}
		}
		listLimit := m.listLimit()

		if len(m.core.Filtered) > 0 {
			rows := make([]suggestionRow, 0, len(m.core.Filtered)+1)
			for _, dir := range m.core.Filtered {
				rows = append(rows, suggestionRow{
					primary: dir.Name,
					detail:  m.displayPath(dir.Path),
				})
			}
			if createLabel != "" {
				rows = append(rows, createRow)
			}
			list, window := m.renderSuggestionList(rows, m.core.SelectedIdx, listLimit)
			count := m.renderCount(window)
			content = input + "\n" + list + count
		} else if createLabel != "" {
			list, window := m.renderSuggestionList([]suggestionRow{createRow}, m.core.SelectedIdx, listLimit)
			count := m.renderCount(window)
			content = input + "\n" + list + count
		} else {
			content = input + "\n" + m.styles.EmptyState.Render("No matches. Press esc to quit.")
		}
		helpLine = m.renderHelpLine([]struct{ key, desc string }{
			{"enter", "select"}, {"ctrl+d", "delete"}, {"ctrl+s", "sessions"}, {"?", "help"}, {"esc", "quit"},
		})

	case core.ModeProjectDeleteConfirm:
		header = m.styles.Title.Render("⚠ Delete Project")
		breadcrumb = m.renderBreadcrumb()
		prompt := m.styles.Body.Render("This will delete the project and all workspaces:")
		path := m.styles.Path.Render("  " + m.displayPath(m.core.ProjectDeletePath))
		warning := m.styles.Body.Render("This action cannot be undone.")
		actions := m.styles.Key.Render("enter") + " " + m.styles.DestructiveAction.Render("delete") + "  " + m.styles.Key.Render("esc") + " " + m.styles.Help.Render("cancel")
		content = prompt + "\n\n" + path + "\n\n" + warning + "\n\n" + actions
		helpLine = ""

	case core.ModeWorktree:
		header = m.styles.Title.Render("Step 2: Select Workspace")
		breadcrumb = m.renderBreadcrumb()
		prompt := m.styles.Prompt.Render("Select workspace or create new branch:")
		input := prompt + " " + m.worktreeInput.View()
		createName, canCreate := m.core.CreateWorktreeName()
		createLabel := ""
		createRow := suggestionRow{}
		if canCreate {
			createLabel = createName
			createRow = suggestionRow{primary: createLabel, actionLabel: "create"}
		}
		listLimit := m.listLimit()

		if len(m.core.FilteredWT) > 0 {
			rows := make([]suggestionRow, 0, len(m.core.FilteredWT)+1)
			for _, wt := range m.core.FilteredWT {
				rows = append(rows, suggestionRow{primary: m.worktreeDisplayLabel(wt)})
			}
			if createLabel != "" {
				rows = append(rows, createRow)
			}
			list, window := m.renderSuggestionList(rows, m.core.WorktreeIdx, listLimit)
			count := m.renderCount(window)
			content = input + "\n" + list + count
		} else if createLabel != "" {
			list, window := m.renderSuggestionList([]suggestionRow{createRow}, m.core.WorktreeIdx, listLimit)
			count := m.renderCount(window)
			content = input + "\n" + list + count
		} else {
			content = input + "\n" + m.styles.EmptyState.Render("No matches. Press esc to go back.")
		}

		if m.core.ProjectWarning != "" {
			content += "\n" + m.styles.Warning.Render("⚠ "+m.core.ProjectWarning)
		}
		if m.core.WorktreeWarning != "" {
			content += "\n" + m.styles.Warning.Render("⚠ "+m.core.WorktreeWarning)
		}
		helpLine = m.renderHelpLine([]struct{ key, desc string }{
			{"enter", "select"}, {"ctrl+d", "delete"}, {"ctrl+s", "sessions"}, {"?", "help"}, {"esc", "back"},
		})

	case core.ModeWorktreeDeleteConfirm:
		header = m.styles.Title.Render("⚠ Delete Workspace")
		breadcrumb = m.renderBreadcrumb()
		labelText := m.worktreeBreadcrumbLabel()
		if labelText == "" {
			labelText = m.displayPath(m.core.WorktreeDeletePath)
		}
		label := m.styles.Body.Render("  " + labelText)
		path := m.styles.Path.Render("  " + m.displayPath(m.core.WorktreeDeletePath))
		prompt := m.styles.Body.Render("This will delete the following workspace:")
		warning := m.styles.Body.Render("This action cannot be undone.")
		actions := m.styles.Key.Render("enter") + " " + m.styles.DestructiveAction.Render("delete") + "  " + m.styles.Key.Render("esc") + " " + m.styles.Help.Render("cancel")
		content = prompt + "\n\n" + label + "\n" + path + "\n\n" + warning + "\n\n" + actions
		helpLine = ""

	case core.ModeTool:
		header = m.styles.Title.Render("Step 3: Select Tool")
		breadcrumb = m.renderBreadcrumb()
		prompt := m.styles.Prompt.Render("Select tool:")
		input := prompt + " " + m.toolInput.View()
		listLimit := m.listLimit()

		if len(m.core.FilteredTools) > 0 {
			rows := make([]suggestionRow, 0, len(m.core.FilteredTools))
			for _, tool := range m.core.FilteredTools {
				rows = append(rows, suggestionRow{primary: tool})
			}
			list, window := m.renderSuggestionList(rows, m.core.ToolIdx, listLimit)
			count := m.renderCount(window)
			content = input + "\n" + list + count
		} else {
			content = input + "\n" + m.styles.EmptyState.Render("No matches. Press esc to go back.")
		}
		if m.core.ToolError != "" {
			content += "\n" + m.styles.Error.Render(m.core.ToolError)
		}
		helpLine = m.renderHelpLine([]struct{ key, desc string }{
			{"enter", "open"}, {"ctrl+s", "sessions"}, {"?", "help"}, {"esc", "back"},
		})

	case core.ModeToolStarting:
		toolName := "tool"
		if m.core.PendingSpec != nil && m.core.PendingSpec.Tool != "" {
			toolName = m.core.PendingSpec.Tool
		}
		breadcrumb = m.renderBreadcrumb()
		content = fmt.Sprintf("%s Starting %s...", m.spinner.View(), toolName)
		helpLine = m.renderHelpLine([]struct{ key, desc string }{{"esc", "back"}})

	case core.ModeSessions:
		header = m.styles.Title.Render("Active tmux sessions")
		breadcrumb = m.renderBreadcrumb()
		if len(m.core.Sessions) == 0 {
			content = m.styles.EmptyState.Render("No active sessions. Press esc to return.")
			helpLine = m.renderHelpLine([]struct{ key, desc string }{{"?", "help"}, {"esc", "back"}})
			break
		}
		prompt := m.styles.Prompt.Render("Filter sessions:")
		input := prompt + " " + m.sessionInput.View()
		rows := make([]suggestionRow, 0, len(m.core.FilteredSessions))
		for _, session := range m.core.FilteredSessions {
			label := core.SessionDisplayLabel(session)
			if label == "" {
				label = m.displayPath(session.DirPath)
			}
			rows = append(rows, suggestionRow{primary: label})
		}
		listLimit := m.listLimit()
		if len(rows) > 0 {
			list, window := m.renderSuggestionList(rows, m.core.SessionIdx, listLimit)
			count := m.renderCount(window)
			content = input + "\n" + list + count
		} else {
			content = input + "\n" + m.styles.EmptyState.Render("No matches. Press esc to return.")
		}
		helpLine = m.renderHelpLine([]struct{ key, desc string }{{"enter", "attach"}, {"?", "help"}, {"esc", "back"}})

	case core.ModeError:
		content = m.styles.Error.Render(fmt.Sprintf("Error: %v", m.core.Err))
		helpLine = m.renderHelpLine([]struct{ key, desc string }{{"esc", "quit"}})
	}

	if header != "" {
		if breadcrumb != "" {
			content = header + "\n" + breadcrumb + "\n\n" + content
		} else {
			content = header + "\n\n" + content
		}
	}

	if helpLine != "" {
		content += "\n\n" + helpLine
	}

	boxStyle := m.styles.BoxWithWidth(m.width)
	box := boxStyle.Render(content)

	if m.height <= 0 || m.width <= 0 {
		return box
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box,
	)
}

func (m Model) renderBreadcrumb() string {
	items := make([]string, 0, 3)
	if m.core.SelectedProject != "" {
		items = append(items, m.renderBreadcrumbItem("Project", filepath.Base(m.core.SelectedProject)))
	}
	if m.core.SelectedWorktreePath != "" {
		items = append(items, m.renderBreadcrumbItem("Workspace", m.worktreeBreadcrumbLabel()))
	}
	if m.core.Mode == core.ModeTool || m.core.Mode == core.ModeToolStarting {
		if tool, ok := m.core.SelectedTool(); ok {
			items = append(items, m.renderBreadcrumbItem("Tool", tool))
		} else if m.core.PendingSpec != nil && m.core.PendingSpec.Tool != "" {
			items = append(items, m.renderBreadcrumbItem("Tool", m.core.PendingSpec.Tool))
		}
	}
	if len(items) == 0 {
		return ""
	}
	return strings.Join(items, m.styles.Help.Render("  •  "))
}

func (m Model) worktreeBreadcrumbLabel() string {
	if m.core.SelectedWorktreePath == "" {
		return ""
	}
	selectedPath := filepath.Clean(m.core.SelectedWorktreePath)
	for _, wt := range m.core.Worktrees {
		if filepath.Clean(wt.Path) == selectedPath {
			if wt.Branch != "" {
				return wt.Branch
			}
			if wt.Name != "" {
				return wt.Name
			}
		}
	}
	return filepath.Base(m.core.SelectedWorktreePath)
}

func (m Model) worktreeDisplayLabel(wt core.Worktree) string {
	if wt.Branch != "" {
		return wt.Branch
	}
	return wt.Name
}

func (m Model) renderBreadcrumbItem(label, value string) string {
	return fmt.Sprintf("%s %s", m.styles.BreadcrumbLabel.Render(label+":"), m.styles.BreadcrumbValue.Render(value))
}

func (m Model) renderCount(window suggestionWindow) string {
	if window.total == 0 || window.end == 0 {
		return ""
	}
	line := fmt.Sprintf("Showing %d-%d of %d", window.start+1, window.end, window.total)
	return "\n" + m.styles.Count.Render(line)
}

func (m Model) renderHelpModal() string {
	header := m.styles.Title.Render("Help Menu")

	sections := []struct {
		title string
		rows  []string
	}{
		{
			title: "Navigation",
			rows: []string{
				m.renderHelpRow("up/down", "move selection"),
				m.renderHelpRow("ctrl+j/ctrl+k", "move selection"),
				m.renderHelpRow("enter", "select / open / create"),
				m.renderHelpRow("esc", "back or quit"),
				m.renderHelpRow("ctrl+c", "quit"),
				m.renderHelpRow("?", "toggle help"),
			},
		},
		{
			title: "Filtering",
			rows: []string{
				m.renderHelpRow("type", "filter the current list"),
			},
		},
		{
			title: "Sessions",
			rows: []string{
				m.renderHelpRow("ctrl+s", "open sessions"),
				m.renderHelpRow("enter", "attach to session"),
			},
		},
		{
			title: "Actions",
			rows: []string{
				m.renderHelpRow("ctrl+d", "delete project/workspace"),
				m.renderHelpRow("ctrl+t", "open theme picker"),
			},
		},
	}

	var out strings.Builder
	for i, section := range sections {
		if i > 0 {
			out.WriteString("\n\n")
		}
		out.WriteString(m.styles.SectionTitle.Render(section.title))
		out.WriteString("\n")
		out.WriteString(strings.Join(section.rows, "\n"))
	}

	footer := m.styles.Help.Render("Press ? or esc to close")
	content := header + "\n\n" + out.String() + "\n\n" + footer
	boxStyle := m.styles.BoxWithWidth(m.width)
	box := boxStyle.Render(content)

	if m.height <= 0 || m.width <= 0 {
		return box
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box,
	)
}

func (m Model) renderThemePicker() string {
	header := m.styles.Title.Render("Theme Picker")
	prompt := m.styles.Prompt.Render("Filter themes:")
	input := prompt + " " + m.themeInput.View()

	rows := make([]suggestionRow, 0, len(m.filteredThemes))
	for _, theme := range m.filteredThemes {
		rows = append(rows, suggestionRow{primary: theme.Name})
	}

	listLimit := m.listLimit()
	var content string
	if len(rows) > 0 {
		list, window := m.renderSuggestionList(rows, m.themeIdx, listLimit)
		count := m.renderCount(window)
		content = input + "\n" + list + count
	} else {
		content = input + "\n" + m.styles.EmptyState.Render("No matching themes.")
	}

	help := m.renderHelpLine([]struct{ key, desc string }{
		{"enter", "apply"}, {"esc", "cancel"},
	})
	content = header + "\n\n" + content + "\n\n" + help

	boxStyle := m.styles.BoxWithWidth(m.width)
	box := boxStyle.Render(content)

	if m.height <= 0 || m.width <= 0 {
		return box
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box,
	)
}
