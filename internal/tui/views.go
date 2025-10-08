package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kanekitakitos/cmd-vault/internal/models"
)

func (m model) renderView() string {
	var mainView string
	const verticalLayoutBreakpoint = 80

	// Stay in file browser view if we are in it, or running a command that was started from it.
	if m.state == stateFileBrowser || m.state == stateRunInPath || (m.state == stateRunningCmd && m.previousState == stateRunInPath) {
		mainView = m.viewFileBrowser()
	} else if m.width < verticalLayoutBreakpoint {
		mainView = m.viewVertical()
	} else {
		mainView = m.viewNormalHorizontal()
	}

	footer := footerStyle.Render(m.getFooterContent())
	mainContent := lipgloss.JoinVertical(lipgloss.Left, mainView, footer)

	overlay := m.viewOverlay()
	if overlay != "" {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
	}

	return mainContent
}

func (m model) viewVertical() string {
	availableHeight := m.height - 4
	// Subtract a margin to prevent panels from touching the window edges
	panelWidth := m.width - 2
	listHeight := availableHeight / 4
	detailsHeight := availableHeight / 4
	outputHeight := availableHeight - listHeight - detailsHeight

	listContent := renderList(m.commands, m.selected, panelWidth-4)
	listPanel := panelStyle.Copy().Width(panelWidth).Height(listHeight).Render(listContent)

	detailsContent := "No commands"
	if len(m.commands) > 0 {
		detailsContent = renderDetails(&m.commands[m.selected]) + "\n" + renderNote(&m.commands[m.selected], panelWidth-2)
	} else if m.state == stateContextHelp {
		detailsContent = renderHelpContent()
	}
	detailsPanel := panelStyle.Copy().Width(panelWidth).Height(detailsHeight).Render(detailsContent)

	outputPanelStyle := panelStyle.Copy().Width(panelWidth).Height(outputHeight)
	if m.state == stateOutputFocus {
		outputPanelStyle = outputPanelStyle.BorderForeground(secondaryColor)
	}
	m.outputViewport.Width = panelWidth - 2
	m.outputViewport.Height = outputHeight - 2
	outputPanel := outputPanelStyle.Render(m.outputViewport.View())

	// Center the vertical layout
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Left, listPanel, detailsPanel, outputPanel))
}

func (m model) viewNormalHorizontal() string {
	mainPanelHeight := m.height - 4
	availableWidth := m.width - 4 // Subtract a margin
	leftPanelWidth := int(float32(availableWidth) * 0.35)
	rightPanelWidth := availableWidth - leftPanelWidth

	leftContent := renderList(m.commands, m.selected, leftPanelWidth-2)
	leftPanel := panelStyle.Copy().
		Width(leftPanelWidth).
		Height(mainPanelHeight).
		Render(leftContent)

	detailsContent := "No commands available."
	if len(m.commands) > 0 {
		c := &m.commands[m.selected]
		detailsContent = lipgloss.JoinVertical(lipgloss.Left, renderDetails(c), renderNote(c, rightPanelWidth-2))
	} else if m.state == stateContextHelp {
		detailsContent = renderHelpContent()
	}

	detailsHeight := mainPanelHeight / 3
	outputHeight := mainPanelHeight - detailsHeight
	detailsPanel := panelStyle.Copy().Width(rightPanelWidth).Height(detailsHeight).Render(detailsContent)

	outputPanelStyle := panelStyle.Copy().Width(rightPanelWidth).Height(outputHeight)
	if m.state == stateOutputFocus {
		outputPanelStyle = outputPanelStyle.BorderForeground(secondaryColor)
	}
	m.outputViewport.Width = rightPanelWidth - 2
	m.outputViewport.Height = outputHeight - 2
	outputPanel := outputPanelStyle.Render(m.outputViewport.View())
	rightPanel := lipgloss.JoinVertical(lipgloss.Left, detailsPanel, outputPanel)

	// Center the horizontal layout
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel))
}

func (m model) viewFileBrowser() string {
	mainPanelHeight := m.height - 4
	leftPanelWidth := int(float32(m.width) * 0.35)
	rightPanelWidth := m.width - 4 - leftPanelWidth

	leftContent := renderFileBrowser(m.files, m.selectedFile, m.currentPath, leftPanelWidth-2)
	leftPanel := panelStyle.Copy().
		Width(leftPanelWidth).
		Height(mainPanelHeight).
		Render(leftContent)

	var selectedEntry os.DirEntry
	if len(m.files) > 0 && m.selectedFile < len(m.files) {
		selectedEntry = m.files[m.selectedFile]
	}
	fileDetailsContent := renderFileBrowserDetails(selectedEntry, rightPanelWidth-2)
	if m.state == stateContextHelp {
		fileDetailsContent = renderHelpContent()
	}

	var fileActionsContent string
	if m.state == stateRunInPath {
		fileActionsContent = "> " + m.runInput.View()
	} else {
		fileActionsContent = "  [Arrows] Navigate   [s] Exit   [r] Run here"
	}

	detailsHeight := (mainPanelHeight / 2) - 2
	outputHeight := mainPanelHeight - detailsHeight - 2

	detailsPanel := panelStyle.Copy().
		Width(rightPanelWidth).
		Height(detailsHeight).
		Render(fileDetailsContent)

	outputPanelStyle := panelStyle.Copy().Width(rightPanelWidth).Height(outputHeight)
	if m.state == stateOutputFocus {
		outputPanelStyle = outputPanelStyle.BorderForeground(secondaryColor)
	}
	m.outputViewport.Width = rightPanelWidth - 2
	m.outputViewport.Height = outputHeight - 2

	var outputContent string
	if m.state == stateRunInPath {
		outputContent = m.outputViewport.View() + "\n" + fileActionsContent
	} else {
		outputContent = m.outputViewport.View()
	}
	outputPanel := outputPanelStyle.Render(outputContent)
	rightPanel := lipgloss.JoinVertical(lipgloss.Left, detailsPanel, outputPanel)
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel))
}

func (m model) viewOverlay() string {
	switch m.state {
	case stateHelp:
		return renderHelpView()
	case stateActionsPanel:
		return renderActionsPanel(m.actions, m.selectedAction)
	case stateAdd, stateEdit:
		title := "Add Command"
		if m.state == stateEdit {
			title = "Edit Command"
		}
		form := lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render(title),
			"Name: "+m.nameInput.View(),
			"Cmd:  "+m.cmdInput.View(),
			"Note: "+m.noteInput.View(),
			"\nPress Enter to save, Esc to cancel",
		)
		return borderStyle.Render(lipgloss.NewStyle().Padding(1).Render(form))
	case stateConfirmDelete:
		return borderStyle.Render(lipgloss.NewStyle().Padding(1).SetString("Confirm delete? (y/n)").String())
	case stateConfirmCancel:
		return borderStyle.Render(lipgloss.NewStyle().Padding(1).SetString("Discard changes? (y/n)").String())
	case stateRunningCmd:
		return "" // No longer an overlay, handled inline
	}
	return ""
}

func (m *model) getFooterContent() string {
	if m.state == stateFileBrowser {
		return "[S] Exit Files  [X] Help  [Q] Quit  " + m.footerMsg
	}
	return "[R] Run  [S] Files  [X] Help  [Q] Quit  " + m.footerMsg
}

func renderList(commands []models.Command, selected int, width int) string {
	var b strings.Builder
	title := titleStyle.Render("Commands")
	b.WriteString(title)
	b.WriteString("\n")
	for i, c := range commands {
		style := lipgloss.NewStyle()
		prefix := "  "
		if i == selected {
			style = style.Foreground(primaryColor).Bold(true)
			prefix = "→ "
		}
		name := c.Name
		usage := fmt.Sprintf("(%d)", c.UsageCount)
		availableWidth := width - len(prefix) - len(usage) - 1
		if len(name) > availableWidth {
			name = name[:availableWidth-3] + "..."
		}
		line := fmt.Sprintf("%s%s %s", prefix, name, usage)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}
	return b.String()
}

func renderDetails(c *models.Command) string {
	if c == nil {
		return "No command selected"
	}
	return fmt.Sprintf("%s\n%s", titleStyle.Render(c.Name), c.CommandStr)
}

func renderNote(c *models.Command, width int) string {
	if c == nil {
		return "No note"
	}
	return lipgloss.NewStyle().Width(width).Render(c.Note)
}

func renderFileBrowser(files []os.DirEntry, selected int, path string, width int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Explorer: "+path) + "\n")
	dirStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	for i, f := range files {
		style := lipgloss.NewStyle()
		prefix := "  "
		if i == selected {
			style = style.Foreground(primaryColor).Bold(true)
			prefix = "→ "
		}
		name := f.Name()
		if f.IsDir() {
			name = dirStyle.Render(name + "/")
		} else {
			name = style.Render(name)
		}
		b.WriteString(fmt.Sprintf("%s%s\n", prefix, name))
	}
	return b.String()
}

func renderFileBrowserDetails(entry os.DirEntry, width int) string {
	if entry == nil {
		return "No file selected."
	}
	info, err := entry.Info()
	if err != nil {
		return fmt.Sprintf("Error reading info: %v", err)
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("Details") + "\n")
	details := []string{
		fmt.Sprintf("Name:      %s", info.Name()),
		fmt.Sprintf("Size:      %d bytes", info.Size()),
		fmt.Sprintf("Mode:      %s", info.Mode().String()),
		fmt.Sprintf("Modified:  %s", info.ModTime().Format("2006-01-02 15:04:05")),
	}
	for _, d := range details {
		b.WriteString(d + "\n")
	}
	return b.String()
}

type helpBinding struct {
	Key         string
	Description string
}

var helpBindings = []helpBinding{
	{Key: "↑/k, ↓/j", Description: "Navigate lists"},
	{Key: "r", Description: "Run command (or enter mini-terminal)"},
	{Key: "s", Description: "Open/close file browser"},
	{Key: "o", Description: "Focus/scroll output panel"},
	{Key: "a, e, d", Description: "Add, Edit, Delete command"},
	{Key: "c", Description: "Copy current path (in browser)"},
	{Key: "p", Description: "Paste saved command (in mini-terminal)"},
	{Key: "x", Description: "Show/hide contextual help"},
	{Key: "q", Description: "Quit program"},
}

func renderHelpContent() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Help") + "\n")
	for _, h := range helpBindings {
		b.WriteString(fmt.Sprintf("  %-12s %s\n", h.Key, h.Description))
	}
	return b.String()
}

func renderHelpView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Help") + "\n")
	for _, h := range helpBindings {
		b.WriteString(fmt.Sprintf("  %-20s %s\n", h.Key, h.Description))
	}
	b.WriteString("\nPress '?' or 'Esc' to close.")
	return borderStyle.Render(b.String())
}

func renderActionsPanel(actions []string, selected int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Actions Panel") + "\n")
	for i, action := range actions {
		style := lipgloss.NewStyle()
		prefix := "  "
		if i == selected {
			style = style.Foreground(primaryColor).Bold(true)
			prefix = "→ "
		}
		line := fmt.Sprintf("%s%s", prefix, action)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}
	b.WriteString("\nUse ↑/↓ to navigate, Enter to select.")
	return borderStyle.Render(lipgloss.NewStyle().Padding(1).Render(b.String()))
}

func renderSelectCmdToPaste(commands []models.Command, selected int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select Command to Paste") + "\n\n")
	for i, c := range commands {
		style := lipgloss.NewStyle()
		prefix := "  "
		if i == selected {
			style = style.Foreground(primaryColor).Bold(true)
			prefix = "→ "
		}
		b.WriteString(style.Render(fmt.Sprintf("%s%s", prefix, c.Name)) + "\n")
	}
	b.WriteString("\nUse ↑/↓ to navigate, Enter to paste, Esc to cancel.")
	return borderStyle.Render(lipgloss.NewStyle().Padding(1).Render(b.String()))
}
