package ui

import "github.com/charmbracelet/lipgloss"

var (
	gruvboxYellow = lipgloss.Color("#c5a97a")
	red           = lipgloss.Color("#e06c75")
	lightGray     = lipgloss.Color("#6a6a6a")

	suggestionStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	navStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	promptStyle = lipgloss.NewStyle().
			Foreground(gruvboxYellow)

	errorStyle = lipgloss.NewStyle().
			Foreground(red)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(gruvboxYellow).
			Padding(0, 4).
			Width(90)
)
