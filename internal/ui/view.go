package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ariguillegp/solo/internal/core"
)

func renderHelpLine(items []struct{ key, desc string }) string {
	var parts []string
	for _, item := range items {
		parts = append(parts, keyStyle.Render(item.key)+" "+helpStyle.Render(item.desc))
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, strings.Join(parts, "  "))
}

const maxSuggestions = 5

func renderSuggestionList(lines []string, selectedIdx int) string {
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
		out.WriteString(scrollIndicatorStyle.Render("  ▲ more above"))
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
			row = selectedSuggestionStyle.Render(prefix + lines[i])
		} else {
			row = suggestionStyle.Render(prefix + lines[i])
		}
		out.WriteString(row)
	}

	if end < len(lines) {
		out.WriteString("\n")
		out.WriteString(scrollIndicatorStyle.Render("  ▼ more below"))
	}

	return out.String()
}

func (m Model) View() string {
	var content string
	var helpLine string
	var header string

	switch m.core.Mode {
	case core.ModeLoading:
		content = m.spinner.View() + " Scanning..."
		helpLine = renderHelpLine([]struct{ key, desc string }{{"esc", "quit"}})

	case core.ModeBrowsing:
		header = titleStyle.Render("Step 1: Select Project")
		prompt := promptStyle.Render("Enter the project directory:")
		input := prompt + " " + m.input.View()

		if len(m.core.Filtered) > 0 {
			lines := make([]string, 0, len(m.core.Filtered))
			for _, dir := range m.core.Filtered {
				lines = append(lines, dir.Path)
			}
			content = input + "\n" + renderSuggestionList(lines, m.core.SelectedIdx)
		} else if m.core.Query != "" {
			content = input + "\n" + suggestionStyle.Render("(create new)")
		} else {
			content = input
		}
		helpLine = renderHelpLine([]struct{ key, desc string }{
			{"↑/↓", "navigate"}, {"enter", "select"}, {"ctrl+n", "create"}, {"esc", "quit"},
		})

	case core.ModeWorktree:
		header = titleStyle.Render("Step 2: Select Worktree")
		prompt := promptStyle.Render("Select worktree or create new branch:")
		input := prompt + " " + m.worktreeInput.View()

		if len(m.core.FilteredWT) > 0 {
			lines := make([]string, 0, len(m.core.FilteredWT))
			for _, wt := range m.core.FilteredWT {
				lines = append(lines, wt.Path)
			}
			content = input + "\n" + renderSuggestionList(lines, m.core.WorktreeIdx)
		} else if m.core.WorktreeQuery != "" {
			content = input + "\n" + suggestionStyle.Render("(create new: "+m.core.WorktreeQuery+")")
		} else {
			content = input
		}

		if m.core.ProjectWarning != "" {
			content += "\n" + warningStyle.Render(m.core.ProjectWarning)
		}
		helpLine = renderHelpLine([]struct{ key, desc string }{
			{"↑/↓", "navigate"}, {"enter", "select"}, {"ctrl+n", "create"}, {"ctrl+d", "delete"}, {"esc", "back"},
		})

	case core.ModeWorktreeDeleteConfirm:
		header = destructiveTitleStyle.Render("⚠ Delete Worktree")
		prompt := destructiveTextStyle.Render("This will delete the following worktree:")
		path := destructiveTextStyle.Render("  " + m.core.WorktreeDeletePath)
		warning := destructiveTextStyle.Render("This action cannot be undone.")
		actions := keyStyle.Render("enter") + " " + destructiveActionStyle.Render("delete") + "  " + keyStyle.Render("esc") + " " + helpStyle.Render("cancel")
		content = prompt + "\n\n" + path + "\n\n" + warning + "\n\n" + actions
		helpLine = ""

	case core.ModeTool:
		header = titleStyle.Render("Step 3: Select Tool")
		prompt := promptStyle.Render("Select tool:")
		input := prompt + " " + m.toolInput.View()

		if len(m.core.FilteredTools) > 0 {
			content = input + "\n" + renderSuggestionList(m.core.FilteredTools, m.core.ToolIdx)
		} else {
			content = input
		}
		helpLine = renderHelpLine([]struct{ key, desc string }{
			{"↑/↓", "navigate"}, {"enter", "open"}, {"esc", "back"},
		})

	case core.ModeError:
		content = errorStyle.Render(fmt.Sprintf("Error: %v", m.core.Err))
		helpLine = renderHelpLine([]struct{ key, desc string }{{"esc", "quit"}})
	}

	if header != "" {
		content = header + "\n\n" + content
	}

	if helpLine != "" {
		content += "\n\n" + helpLine
	}

	boxStyle := boxStyleWithWidth(m.width)
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
