package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors requested: amber/neon on black
	bgColor        = lipgloss.Color("#000000")
	primaryColor   = lipgloss.Color("#FFBF00") // amber neon
	secondaryColor = lipgloss.Color("#E5AA00")

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1)
	panelStyle = lipgloss.NewStyle().
			Background(bgColor).
			Foreground(primaryColor).
			Padding(0, 1)
	footerStyle = lipgloss.NewStyle().
			Background(bgColor).
			Foreground(primaryColor).
			Align(lipgloss.Center).
			Padding(0, 1)
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)
)
