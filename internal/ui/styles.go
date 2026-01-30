package ui

import "github.com/charmbracelet/lipgloss"

var (
	gruvboxYellow = lipgloss.Color("#c5a97a")
	red           = lipgloss.Color("#e06c75")
	lightGray     = lipgloss.Color("#6a6a6a")
	white         = lipgloss.Color("#ffffff")
	warningOrange = lipgloss.Color("#e0b36c")

	suggestionStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	selectedStyle = lipgloss.NewStyle().
			Foreground(white).
			Bold(true)

	navStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	promptStyle = lipgloss.NewStyle().
			Foreground(gruvboxYellow)

	errorStyle = lipgloss.NewStyle().
			Foreground(red)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningOrange)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(gruvboxYellow).
			Padding(0, 4).
			Width(90)
)
