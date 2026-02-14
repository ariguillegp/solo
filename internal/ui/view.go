package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"github.com/ariguillegp/rivet/internal/core"
)

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
		helpLine = m.shortHelpView()

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
		helpLine = m.shortHelpView()

	case core.ModeProjectDeleteConfirm:
		header = m.styles.Title.Render("⚠ Delete Project")
		breadcrumb = m.renderBreadcrumb()
		prompt := m.styles.Body.Render("This will delete the project and all workspaces:")
		path := m.styles.Path.Render("  " + m.displayPath(m.core.ProjectDeletePath))
		warning := m.styles.Body.Render("This action cannot be undone.")
		actions := m.styles.Key.Render("enter") + " " + m.styles.DestructiveAction.Render("delete") + "  " + m.styles.Key.Render("esc") + " " + m.styles.Help.Render("cancel")
		content = m.renderViewportContent(prompt + "\n\n" + path + "\n\n" + warning + "\n\n" + actions)

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
		helpLine = m.shortHelpView()

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
		content = m.renderViewportContent(prompt + "\n\n" + label + "\n" + path + "\n\n" + warning + "\n\n" + actions)

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
		helpLine = m.shortHelpView()

	case core.ModeToolStarting:
		toolName := "tool"
		if m.core.PendingSpec != nil && m.core.PendingSpec.Tool != "" {
			toolName = m.core.PendingSpec.Tool
		}
		breadcrumb = m.renderBreadcrumb()
		progressValue := m.toolStartingProgress()
		bar := m.progress.ViewAs(progressValue)
		content = fmt.Sprintf("Starting %s...\n\n%s", toolName, bar)
		helpLine = m.shortHelpView()

	case core.ModeSessions:
		header = m.styles.Title.Render("Active tmux sessions")
		breadcrumb = m.renderBreadcrumb()
		if len(m.core.Sessions) == 0 {
			content = m.styles.EmptyState.Render("No active sessions. Press esc to return.")
			helpLine = m.sessionsEmptyShortHelpView()
			break
		}
		prompt := m.styles.Prompt.Render("Filter sessions:")
		input := prompt + " " + m.sessionInput.View()
		if len(m.core.FilteredSessions) == 0 {
			content = input + "\n" + m.styles.EmptyState.Render("No matches. Press esc to return.")
			helpLine = m.shortHelpView()
			break
		}
		if m.sessionListIsCompact() {
			content = input + "\n" + m.sessionList.View() + m.renderCount(m.sessionList)
		} else {
			tableView := m.styles.Path.Render(m.sessionTable.View())
			content = input + "\n" + tableView + m.renderTableCount(m.sessionTable)
		}
		helpLine = m.shortHelpView()

	case core.ModeError:
		errContent := m.styles.Error.Render(fmt.Sprintf("Error: %v", m.core.Err))
		content = m.renderViewportContent(errContent)
		helpLine = m.shortHelpView()
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

	box := m.renderModalBox(content, m.isViewportActive())
	if m.height <= 0 || m.width <= 0 {
		return box
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderTableCount(t table.Model) string {
	total := len(t.Rows())
	if total == 0 {
		return ""
	}
	start, end := visibleListWindow(total, t.Cursor(), t.Height())
	if end == 0 {
		return ""
	}
	line := fmt.Sprintf("Showing %d-%d of %d", start+1, end, total)
	return "\n" + m.styles.Count.Render(line)
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

const maxDisplayedToolStartingProgress = 0.99

func (m Model) displayableToolStartingProgress(value float64) float64 {
	progress := clamp01(value)
	if m.core.Mode == core.ModeToolStarting && progress >= 1 {
		return maxDisplayedToolStartingProgress
	}
	return progress
}

func (m Model) toolStartingProgress() float64 {
	if m.toolStartingAt.IsZero() || m.toolStartingDuration <= 0 {
		return m.displayableToolStartingProgress(0)
	}
	progress := float64(time.Since(m.toolStartingAt)) / float64(m.toolStartingDuration)
	return m.displayableToolStartingProgress(progress)
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
	body := m.fullHelpView() + "\n\n" + m.styles.Help.Render("Press ? or esc to close")
	content := header + "\n\n" + m.renderViewportContent(body)
	box := m.renderModalBox(content, true)
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

	help := m.help.ShortHelpView([]key.Binding{m.keymap.binding(m.keymap.Select, "apply"), m.keymap.binding(m.keymap.Back, "cancel")})
	content = header + "\n\n" + content + "\n\n" + help

	box := m.renderModalBox(content, false)
	if m.height <= 0 || m.width <= 0 {
		return box
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) modalBoxDimensions() (int, int) {
	boxStyle := m.styles.BoxWithWidth(m.width)
	boxWidth := boxStyle.GetWidth()
	if boxWidth <= 0 {
		boxWidth = minBoxWidth
	}
	if m.width > 0 && boxWidth > m.width {
		boxWidth = m.width
	}
	if m.height <= 0 {
		return boxWidth, 0
	}
	boxHeight := m.height * 3 / 4
	if boxHeight < 8 {
		boxHeight = m.height
	}
	if boxHeight > m.height-2 {
		boxHeight = m.height - 2
	}
	if boxHeight < 1 {
		boxHeight = 1
	}
	return boxWidth, boxHeight
}

func (m Model) renderModalBox(content string, fixedHeight bool) string {
	boxStyle := m.styles.BoxWithWidth(m.width)
	if fixedHeight {
		_, boxHeight := m.modalBoxDimensions()
		if boxHeight > 0 {
			boxStyle = boxStyle.Height(boxHeight)
		}
	}
	return boxStyle.Render(content)
}

func (m Model) renderViewportContent(content string) string {
	vp := m.viewport
	boxWidth, boxHeight := m.modalBoxDimensions()
	boxStyle := m.styles.BoxWithWidth(m.width)
	maxWidth := boxWidth - boxStyle.GetHorizontalFrameSize()
	maxHeight := boxHeight - boxStyle.GetVerticalFrameSize()
	if maxWidth < 1 {
		maxWidth = 1
	}
	if maxHeight < 1 {
		maxHeight = 1
	}

	vp.Width = maxWidth
	contentHeight := lipgloss.Height(content)
	if contentHeight < 1 {
		contentHeight = 1
	}
	if contentHeight < maxHeight {
		vp.Height = contentHeight
	} else {
		vp.Height = maxHeight
	}

	vp.SetContent(content)
	return vp.View()
}
