package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

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

func (m Model) View() string {
	m.syncLists()
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
		if len(m.projectList.Items()) > 0 {
			content = input + "\n" + m.projectList.View() + m.renderCount(m.projectList)
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

	case core.ModeWorktree:
		header = m.styles.Title.Render("Step 2: Select Workspace")
		breadcrumb = m.renderBreadcrumb()
		prompt := m.styles.Prompt.Render("Select workspace or create new branch:")
		input := prompt + " " + m.worktreeInput.View()
		if len(m.worktreeList.Items()) > 0 {
			content = input + "\n" + m.worktreeList.View() + m.renderCount(m.worktreeList)
		} else {
			content = input + "\n" + m.styles.EmptyState.Render("No matches. Press esc to go back.")
		}
		if m.core.ProjectWarning != "" {
			content += "\n" + m.styles.Warning.Render("⚠ "+m.core.ProjectWarning)
		}
		if m.core.WorktreeWarning != "" {
			content += "\n" + m.styles.Warning.Render("⚠ "+m.core.WorktreeWarning)
		}
		helpLine = m.renderHelpLine([]struct{ key, desc string }{{"enter", "select"}, {"ctrl+d", "delete"}, {"ctrl+s", "sessions"}, {"?", "help"}, {"esc", "back"}})

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

	case core.ModeTool:
		header = m.styles.Title.Render("Step 3: Select Tool")
		breadcrumb = m.renderBreadcrumb()
		prompt := m.styles.Prompt.Render("Select tool:")
		input := prompt + " " + m.toolInput.View()
		if len(m.toolList.Items()) > 0 {
			content = input + "\n" + m.toolList.View() + m.renderCount(m.toolList)
		} else {
			content = input + "\n" + m.styles.EmptyState.Render("No matches. Press esc to go back.")
		}
		if m.core.ToolError != "" {
			content += "\n" + m.styles.Error.Render(m.core.ToolError)
		}
		helpLine = m.renderHelpLine([]struct{ key, desc string }{{"enter", "open"}, {"ctrl+s", "sessions"}, {"?", "help"}, {"esc", "back"}})

	case core.ModeToolStarting:
		toolName := "tool"
		if m.core.PendingSpec != nil && m.core.PendingSpec.Tool != "" {
			toolName = m.core.PendingSpec.Tool
		}
		breadcrumb = m.renderBreadcrumb()
		progressValue := m.toolStartingProgress()
		bar := m.progress.ViewAs(progressValue)
		content = fmt.Sprintf("Starting %s...\n\n%s", toolName, bar)
		helpLine = m.renderHelpLine([]struct{ key, desc string }{{"esc", "cancel"}, {"ctrl+c", "quit"}})

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
		if len(m.sessionList.Items()) > 0 {
			content = input + "\n" + m.sessionList.View() + m.renderCount(m.sessionList)
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

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func (m Model) toolStartingProgress() float64 {
	total := m.core.ToolWarmupTotal
	if total <= 0 {
		return 1
	}

	completed := m.core.ToolWarmupCompleted
	checksProgress := float64(completed) / float64(total)

	if m.core.PendingSpec == nil || m.core.ToolWarmStart == nil {
		return clamp01(checksProgress)
	}

	start, ok := m.core.ToolWarmStart[m.core.PendingSpec.Tool]
	if !ok {
		return clamp01(checksProgress)
	}
	if start.IsZero() {
		return 1
	}

	elapsedFraction := float64(time.Since(start)) / float64(toolReadyDelay)
	elapsedFraction = clamp01(elapsedFraction)

	adjustedCompleted := completed - 1
	if adjustedCompleted < 0 {
		adjustedCompleted = 0
	}
	return clamp01((float64(adjustedCompleted) + elapsedFraction) / float64(total))
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

func (m Model) renderHelpModal() string {
	header := m.styles.Title.Render("Help Menu")

	sections := []struct {
		title string
		rows  []string
	}{
		{title: "Navigation", rows: []string{
			m.renderHelpRow("up/down", "move selection"),
			m.renderHelpRow("ctrl+j/ctrl+k", "move selection"),
			m.renderHelpRow("enter", "select / open / create"),
			m.renderHelpRow("esc", "back or quit"),
			m.renderHelpRow("ctrl+c", "quit"),
			m.renderHelpRow("?", "toggle help"),
		}},
		{title: "Filtering", rows: []string{m.renderHelpRow("type", "filter the current list")}},
		{title: "Sessions", rows: []string{m.renderHelpRow("ctrl+s", "open sessions"), m.renderHelpRow("enter", "attach to session")}},
		{title: "Actions", rows: []string{m.renderHelpRow("ctrl+d", "delete project/workspace"), m.renderHelpRow("ctrl+t", "open theme picker")}},
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
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderThemePicker() string {
	header := m.styles.Title.Render("Theme Picker")
	prompt := m.styles.Prompt.Render("Filter themes:")
	input := prompt + " " + m.themeInput.View()

	var content string
	if len(m.themeList.Items()) > 0 {
		content = input + "\n" + m.themeList.View() + m.renderCount(m.themeList)
	} else {
		content = input + "\n" + m.styles.EmptyState.Render("No matching themes.")
	}

	help := m.renderHelpLine([]struct{ key, desc string }{{"enter", "apply"}, {"esc", "cancel"}})
	content = header + "\n\n" + content + "\n\n" + help

	boxStyle := m.styles.BoxWithWidth(m.width)
	box := boxStyle.Render(content)
	if m.height <= 0 || m.width <= 0 {
		return box
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
