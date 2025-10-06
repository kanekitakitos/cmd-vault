package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Classic 80s terminal green on black
	bgColor        = lipgloss.Color("#000000")
	primaryColor   = lipgloss.Color("#32CD32") // LimeGreen
	secondaryColor = lipgloss.Color("#28a428")

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1)
	// panelStyle is for the main content panels, with borders
	panelStyle = borderStyle.Copy()

	footerStyle = lipgloss.NewStyle().
			Background(bgColor).
			Foreground(primaryColor).
			Align(lipgloss.Center).
			Padding(0, 1)
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)
)
