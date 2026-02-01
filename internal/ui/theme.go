package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name       string
	Accent     lipgloss.Color
	Error      lipgloss.Color
	Warning    lipgloss.Color
	Muted      lipgloss.Color
	Text       lipgloss.Color
	Background lipgloss.Color
}

var themes = []Theme{
	{
		Name:       "Gruvbox",
		Accent:     lipgloss.Color("#c5a97a"),
		Error:      lipgloss.Color("#e06c75"),
		Warning:    lipgloss.Color("#e0b36c"),
		Muted:      lipgloss.Color("#6a6a6a"),
		Text:       lipgloss.Color("#ffffff"),
		Background: lipgloss.Color("#3a3a3a"),
	},
	{
		Name:       "Catppuccin",
		Accent:     lipgloss.Color("#cba6f7"),
		Error:      lipgloss.Color("#f38ba8"),
		Warning:    lipgloss.Color("#fab387"),
		Muted:      lipgloss.Color("#6c7086"),
		Text:       lipgloss.Color("#cdd6f4"),
		Background: lipgloss.Color("#313244"),
	},
	{
		Name:       "Tokyo Night",
		Accent:     lipgloss.Color("#7aa2f7"),
		Error:      lipgloss.Color("#f7768e"),
		Warning:    lipgloss.Color("#e0af68"),
		Muted:      lipgloss.Color("#565f89"),
		Text:       lipgloss.Color("#c0caf5"),
		Background: lipgloss.Color("#24283b"),
	},
	{
		Name:       "Nord",
		Accent:     lipgloss.Color("#88c0d0"),
		Error:      lipgloss.Color("#bf616a"),
		Warning:    lipgloss.Color("#ebcb8b"),
		Muted:      lipgloss.Color("#4c566a"),
		Text:       lipgloss.Color("#eceff4"),
		Background: lipgloss.Color("#3b4252"),
	},
	{
		Name:       "Dracula",
		Accent:     lipgloss.Color("#bd93f9"),
		Error:      lipgloss.Color("#ff5555"),
		Warning:    lipgloss.Color("#ffb86c"),
		Muted:      lipgloss.Color("#6272a4"),
		Text:       lipgloss.Color("#f8f8f2"),
		Background: lipgloss.Color("#44475a"),
	},
}

func Themes() []Theme {
	return themes
}

func ThemeNames() []string {
	names := make([]string, len(themes))
	for i, t := range themes {
		names[i] = t.Name
	}
	return names
}
