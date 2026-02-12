package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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
		Warning:    lipgloss.Color("#f28b2b"),
		Muted:      lipgloss.Color("#6a6a6a"),
		Text:       lipgloss.Color("#ffffff"),
		Background: lipgloss.Color("#3a3a3a"),
	},
	{
		Name:       "Catppuccin",
		Accent:     lipgloss.Color("#cba6f7"),
		Error:      lipgloss.Color("#f38ba8"),
		Warning:    lipgloss.Color("#f28c2a"),
		Muted:      lipgloss.Color("#6c7086"),
		Text:       lipgloss.Color("#cdd6f4"),
		Background: lipgloss.Color("#313244"),
	},
	{
		Name:       "Tokyo Night",
		Accent:     lipgloss.Color("#7aa2f7"),
		Error:      lipgloss.Color("#f7768e"),
		Warning:    lipgloss.Color("#f28b2b"),
		Muted:      lipgloss.Color("#565f89"),
		Text:       lipgloss.Color("#c0caf5"),
		Background: lipgloss.Color("#24283b"),
	},
	{
		Name:       "Nord",
		Accent:     lipgloss.Color("#88c0d0"),
		Error:      lipgloss.Color("#bf616a"),
		Warning:    lipgloss.Color("#f28b2b"),
		Muted:      lipgloss.Color("#4c566a"),
		Text:       lipgloss.Color("#eceff4"),
		Background: lipgloss.Color("#3b4252"),
	},
	{
		Name:       "Dracula",
		Accent:     lipgloss.Color("#bd93f9"),
		Error:      lipgloss.Color("#ff5555"),
		Warning:    lipgloss.Color("#ff8c1a"),
		Muted:      lipgloss.Color("#6272a4"),
		Text:       lipgloss.Color("#f8f8f2"),
		Background: lipgloss.Color("#44475a"),
	},
	{
		Name:       "Solarized Dark",
		Accent:     lipgloss.Color("#268bd2"),
		Error:      lipgloss.Color("#dc322f"),
		Warning:    lipgloss.Color("#b58900"),
		Muted:      lipgloss.Color("#586e75"),
		Text:       lipgloss.Color("#eee8d5"),
		Background: lipgloss.Color("#002b36"),
	},
	{
		Name:       "One Dark",
		Accent:     lipgloss.Color("#61afef"),
		Error:      lipgloss.Color("#e06c75"),
		Warning:    lipgloss.Color("#e5c07b"),
		Muted:      lipgloss.Color("#5c6370"),
		Text:       lipgloss.Color("#abb2bf"),
		Background: lipgloss.Color("#282c34"),
	},
	{
		Name:       "Monokai",
		Accent:     lipgloss.Color("#a6e22e"),
		Error:      lipgloss.Color("#f92672"),
		Warning:    lipgloss.Color("#fd971f"),
		Muted:      lipgloss.Color("#75715e"),
		Text:       lipgloss.Color("#f8f8f2"),
		Background: lipgloss.Color("#272822"),
	},
	{
		Name:       "Rose Pine",
		Accent:     lipgloss.Color("#c4a7e7"),
		Error:      lipgloss.Color("#eb6f92"),
		Warning:    lipgloss.Color("#f6c177"),
		Muted:      lipgloss.Color("#6e6a86"),
		Text:       lipgloss.Color("#e0def4"),
		Background: lipgloss.Color("#191724"),
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

func FilterThemes(themes []Theme, query string) []Theme {
	if query == "" {
		return themes
	}
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return themes
	}

	var result []Theme
	for _, theme := range themes {
		name := strings.ToLower(theme.Name)
		if fuzzyMatch(name, query) {
			result = append(result, theme)
		}
	}
	return result
}

func fuzzyMatch(text, pattern string) bool {
	pi := 0
	for ti := 0; ti < len(text) && pi < len(pattern); ti++ {
		if text[ti] == pattern[pi] {
			pi++
		}
	}
	return pi == len(pattern)
}
