package ui

import "github.com/charmbracelet/lipgloss"

const (
	minBoxWidth     = 60
	maxBoxWidth     = 100
	boxWidthPercent = 50
)

type Styles struct {
	Title              lipgloss.Style
	Suggestion         lipgloss.Style
	SelectedSuggestion lipgloss.Style
	ScrollIndicator    lipgloss.Style
	Help               lipgloss.Style
	Key                lipgloss.Style
	Prompt             lipgloss.Style
	Error              lipgloss.Style
	Warning            lipgloss.Style
	DestructiveTitle   lipgloss.Style
	DestructiveText    lipgloss.Style
	DestructiveAction  lipgloss.Style
	BaseBox            lipgloss.Style
}

func NewStyles(theme Theme) Styles {
	destructiveRed := lipgloss.Color("#ff5555")
	return Styles{
		Title: lipgloss.NewStyle().
			Foreground(theme.Accent).
			Bold(true),
		Suggestion: lipgloss.NewStyle().
			Foreground(theme.Muted),
		SelectedSuggestion: lipgloss.NewStyle().
			Foreground(theme.Text).
			Bold(true),
		ScrollIndicator: lipgloss.NewStyle().
			Foreground(theme.Muted),
		Help: lipgloss.NewStyle().
			Foreground(theme.Muted),
		Key: lipgloss.NewStyle().
			Foreground(theme.Accent).
			Background(theme.Background).
			Padding(0, 1).
			Bold(true),
		Prompt: lipgloss.NewStyle().
			Foreground(theme.Accent),
		Error: lipgloss.NewStyle().
			Foreground(theme.Error),
		Warning: lipgloss.NewStyle().
			Foreground(theme.Warning),
		DestructiveTitle: lipgloss.NewStyle().
			Foreground(destructiveRed).
			Bold(true),
		DestructiveText: lipgloss.NewStyle().
			Foreground(destructiveRed),
		DestructiveAction: lipgloss.NewStyle().
			Foreground(theme.Text).
			Background(destructiveRed).
			Bold(true).
			Padding(0, 1),
		BaseBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Accent).
			Padding(0, 4),
	}
}

func (s Styles) BoxWithWidth(terminalWidth int) lipgloss.Style {
	width := terminalWidth * boxWidthPercent / 100
	if width < minBoxWidth {
		width = minBoxWidth
	}
	if width > maxBoxWidth {
		width = maxBoxWidth
	}
	return s.BaseBox.Width(width)
}
