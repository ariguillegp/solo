package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/ariguillegp/solo/internal/core"
)

func (m Model) View() string {
	var content string

	switch m.core.Mode {
	case core.ModeLoading:
		content = m.spinner.View() + " Scanning..."

	case core.ModeBrowsing:
		prompt := promptStyle.Render("Enter the project directory")
		input := prompt + "\n" + m.input.View()

		if dir, ok := m.core.SelectedDir(); ok {
			suggestion := suggestionStyle.Render(dir.Path)
			nav := ""
			if len(m.core.Filtered) > 1 {
				nav = navStyle.Render(fmt.Sprintf("  [%d/%d]", m.core.SelectedIdx+1, len(m.core.Filtered)))
			}
			content = input + "\n" + suggestion + nav
		} else if m.core.Query != "" {
			content = input + "\n" + suggestionStyle.Render("(create new)")
		} else {
			content = input
		}

	case core.ModeWorktree:
		prompt := promptStyle.Render("Select worktree or create new branch")
		input := prompt + "\n" + m.worktreeInput.View()

		if wt, ok := m.core.SelectedWorktree(); ok {
			suggestion := suggestionStyle.Render(wt.Path)
			nav := ""
			if len(m.core.FilteredWT) > 1 {
				nav = navStyle.Render(fmt.Sprintf("  [%d/%d]", m.core.WorktreeIdx+1, len(m.core.FilteredWT)))
			}
			content = input + "\n" + suggestion + nav
		} else if m.core.WorktreeQuery != "" {
			content = input + "\n" + suggestionStyle.Render("(create new: "+m.core.WorktreeQuery+")")
		} else {
			content = input
		}

		if m.core.ProjectWarning != "" {
			content += "\n" + warningStyle.Render(m.core.ProjectWarning)
		}

	case core.ModeWorktreeDeleteConfirm:
		prompt := promptStyle.Render("Delete worktree?")
		content = prompt + "\n" + suggestionStyle.Render(m.core.WorktreeDeletePath) + "\n" + suggestionStyle.Render("(enter to confirm, esc to cancel)")

	case core.ModeTool:
		prompt := promptStyle.Render("Select tool")
		input := prompt + "\n" + m.toolInput.View()

		if tool, ok := m.core.SelectedTool(); ok {
			suggestion := suggestionStyle.Render(tool)
			nav := ""
			if len(m.core.FilteredTools) > 1 {
				nav = navStyle.Render(fmt.Sprintf("  [%d/%d]", m.core.ToolIdx+1, len(m.core.FilteredTools)))
			}
			content = input + "\n" + suggestion + nav
		} else {
			content = input
		}

	case core.ModeError:
		content = errorStyle.Render(fmt.Sprintf("Error: %v", m.core.Err))
	}

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
