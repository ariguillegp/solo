package ui

import "github.com/charmbracelet/lipgloss"

const (
	minBoxWidth     = 60
	maxBoxWidth     = 100
	boxWidthPercent = 50
)

var (
	gruvboxYellow = lipgloss.Color("#c5a97a")
	red           = lipgloss.Color("#e06c75")
	lightGray     = lipgloss.Color("#6a6a6a")
	warningOrange = lipgloss.Color("#e0b36c")
	white         = lipgloss.Color("#ffffff")

	titleStyle = lipgloss.NewStyle().
			Foreground(gruvboxYellow).
			Bold(true)

	suggestionStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	selectedSuggestionStyle = lipgloss.NewStyle().
				Foreground(white).
				Bold(true)

	scrollIndicatorStyle = lipgloss.NewStyle().
				Foreground(lightGray)

	helpStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	keyStyle = lipgloss.NewStyle().
			Foreground(gruvboxYellow).
			Background(lipgloss.Color("#3a3a3a")).
			Padding(0, 1).
			Bold(true)

	promptStyle = lipgloss.NewStyle().
			Foreground(gruvboxYellow)

	errorStyle = lipgloss.NewStyle().
			Foreground(red)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningOrange)

	destructiveRed = lipgloss.Color("#ff5555")

	destructiveTitleStyle = lipgloss.NewStyle().
				Foreground(destructiveRed).
				Bold(true)

	destructiveTextStyle = lipgloss.NewStyle().
				Foreground(destructiveRed)

	destructiveActionStyle = lipgloss.NewStyle().
				Foreground(white).
				Background(destructiveRed).
				Bold(true).
				Padding(0, 1)

	baseBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(gruvboxYellow).
			Padding(0, 4)
)

func boxStyleWithWidth(terminalWidth int) lipgloss.Style {
	width := terminalWidth * boxWidthPercent / 100
	if width < minBoxWidth {
		width = minBoxWidth
	}
	if width > maxBoxWidth {
		width = maxBoxWidth
	}
	return baseBoxStyle.Width(width)
}
