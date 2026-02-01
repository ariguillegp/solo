package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ariguillegp/solo/internal/core"
)

func (m Model) renderHelpLine(items []struct{ key, desc string }) string {
	var parts []string
	for _, item := range items {
		parts = append(parts, m.styles.Key.Render(item.key)+" "+m.styles.Help.Render(item.desc))
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, strings.Join(parts, "  "))
}

const maxSuggestions = 5

func (m Model) renderSuggestionList(lines []string, selectedIdx int) string {
	if len(lines) == 0 {
		return ""
	}

	start := 0
	if selectedIdx >= maxSuggestions {
		start = selectedIdx - maxSuggestions + 1
	}
	end := start + maxSuggestions
	if end > len(lines) {
		end = len(lines)
		if end-start < maxSuggestions && start > 0 {
			start = end - maxSuggestions
			if start < 0 {
				start = 0
			}
		}
	}

	var out strings.Builder

	if start > 0 {
		out.WriteString(m.styles.ScrollIndicator.Render("  ▲ more above"))
		out.WriteString("\n")
	}

	for i := start; i < end; i++ {
		if i > start || start > 0 {
			out.WriteString("\n")
		}

		prefix := "  "
		if i == selectedIdx {
			prefix = "> "
		}

		var row string
		if i == selectedIdx {
			row = m.styles.SelectedSuggestion.Render(prefix + lines[i])
		} else {
			row = m.styles.Suggestion.Render(prefix + lines[i])
		}
		out.WriteString(row)
	}

	if end < len(lines) {
		out.WriteString("\n")
		out.WriteString(m.styles.ScrollIndicator.Render("  ▼ more below"))
	}

	return out.String()
}

func (m Model) View() string {
	if m.showThemePicker {
		return m.renderThemePicker()
	}

	var content string
	var helpLine string
	var header string

	switch m.core.Mode {
	case core.ModeLoading:
		content = m.spinner.View() + " Scanning..."
		helpLine = m.renderHelpLine([]struct{ key, desc string }{{"esc", "quit"}})

	case core.ModeBrowsing:
		header = m.styles.Title.Render("Step 1: Select Project")
		prompt := m.styles.Prompt.Render("Enter the project directory:")
		input := prompt + " " + m.input.View()

		if len(m.core.Filtered) > 0 {
			lines := make([]string, 0, len(m.core.Filtered))
			for _, dir := range m.core.Filtered {
				lines = append(lines, dir.Path)
			}
			content = input + "\n" + m.renderSuggestionList(lines, m.core.SelectedIdx)
		} else if m.core.Query != "" {
			content = input + "\n" + m.styles.Suggestion.Render("(create new)")
		} else {
			content = input
		}
		helpLine = m.renderHelpLine([]struct{ key, desc string }{
			{"↑/↓", "navigate"}, {"enter", "select"}, {"ctrl+n", "create"}, {"ctrl+t", "theme"}, {"esc", "quit"},
		})

	case core.ModeWorktree:
		header = m.styles.Title.Render("Step 2: Select Worktree")
		prompt := m.styles.Prompt.Render("Select worktree or create new branch:")
		input := prompt + " " + m.worktreeInput.View()

		if len(m.core.FilteredWT) > 0 {
			lines := make([]string, 0, len(m.core.FilteredWT))
			for _, wt := range m.core.FilteredWT {
				lines = append(lines, wt.Path)
			}
			content = input + "\n" + m.renderSuggestionList(lines, m.core.WorktreeIdx)
		} else if m.core.WorktreeQuery != "" {
			content = input + "\n" + m.styles.Suggestion.Render("(create new: "+m.core.WorktreeQuery+")")
		} else {
			content = input
		}

		if m.core.ProjectWarning != "" {
			content += "\n" + m.styles.Warning.Render(m.core.ProjectWarning)
		}
		helpLine = m.renderHelpLine([]struct{ key, desc string }{
			{"↑/↓", "navigate"}, {"enter", "select"}, {"ctrl+n", "create"}, {"ctrl+d", "delete"}, {"ctrl+t", "theme"}, {"esc", "back"},
		})

	case core.ModeWorktreeDeleteConfirm:
		header = m.styles.DestructiveTitle.Render("⚠ Delete Worktree")
		prompt := m.styles.DestructiveText.Render("This will delete the following worktree:")
		path := m.styles.DestructiveText.Render("  " + m.core.WorktreeDeletePath)
		warning := m.styles.DestructiveText.Render("This action cannot be undone.")
		actions := m.styles.Key.Render("enter") + " " + m.styles.DestructiveAction.Render("delete") + "  " + m.styles.Key.Render("esc") + " " + m.styles.Help.Render("cancel")
		content = prompt + "\n\n" + path + "\n\n" + warning + "\n\n" + actions
		helpLine = ""

	case core.ModeTool:
		header = m.styles.Title.Render("Step 3: Select Tool")
		prompt := m.styles.Prompt.Render("Select tool:")
		input := prompt + " " + m.toolInput.View()

		if len(m.core.FilteredTools) > 0 {
			content = input + "\n" + m.renderSuggestionList(m.core.FilteredTools, m.core.ToolIdx)
		} else {
			content = input
		}
		helpLine = m.renderHelpLine([]struct{ key, desc string }{
			{"↑/↓", "navigate"}, {"enter", "open"}, {"ctrl+t", "theme"}, {"esc", "back"},
		})

	case core.ModeError:
		content = m.styles.Error.Render(fmt.Sprintf("Error: %v", m.core.Err))
		helpLine = m.renderHelpLine([]struct{ key, desc string }{{"esc", "quit"}})
	}

	if header != "" {
		content = header + "\n\n" + content
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

func (m Model) renderThemePicker() string {
	header := m.styles.Title.Render("Select Theme")

	var out strings.Builder
	for i, theme := range m.themes {
		prefix := "  "
		if i == m.themeIdx {
			prefix = "> "
		}
		var row string
		if i == m.themeIdx {
			row = m.styles.SelectedSuggestion.Render(prefix + theme.Name)
		} else {
			row = m.styles.Suggestion.Render(prefix + theme.Name)
		}
		if i > 0 {
			out.WriteString("\n")
		}
		out.WriteString(row)
	}

	helpLine := m.renderHelpLine([]struct{ key, desc string }{
		{"↑/↓", "navigate"}, {"enter", "select"}, {"esc", "cancel"},
	})

	content := header + "\n\n" + out.String() + "\n\n" + helpLine

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
