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

	if m.state == stateFileBrowser {
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
	panelWidth := m.width - 4
	listHeight := availableHeight / 2
	detailsHeight := availableHeight - listHeight

	listContent := renderList(m.commands, m.selected, panelWidth-4)
	listPanel := panelStyle.Copy().Width(panelWidth).Height(listHeight).Render(listContent)

	var detailsContent string
	if len(m.commands) > 0 {
		detailsContent = renderDetails(&m.commands[m.selected]) + "\n" + renderNote(&m.commands[m.selected], panelWidth-2)
	} else {
		detailsContent = "No commands"
	}
	detailsPanel := panelStyle.Copy().Width(panelWidth).Height(detailsHeight).Render(detailsContent)

	// Center the vertical layout
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Left, listPanel, detailsPanel))
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

	var rightTopContent, rightBottomContent string
	if len(m.commands) > 0 {
		c := &m.commands[m.selected]
		rightTopContent = renderDetails(c)
		rightBottomContent = renderNote(c, rightPanelWidth-2)
	} else {
		rightTopContent = "No commands"
		rightBottomContent = "No notes"
	}
	rightTopPanel := panelStyle.Copy().Width(rightPanelWidth).Height(4).Render(rightTopContent)
	rightBottomPanel := panelStyle.Copy().Width(rightPanelWidth).Height(mainPanelHeight - 4).Render(rightBottomContent)
	rightPanel := lipgloss.JoinVertical(lipgloss.Left, rightTopPanel, rightBottomPanel)

	// Center the horizontal layout
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel))
}

func (m model) viewFileBrowser() string {
	mainPanelHeight := m.height - 4
	leftPanelWidth := int(float32(m.width) * 0.35)
	rightPanelWidth := m.width - leftPanelWidth

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
	fileActionsContent := "  [Setas] Navegar   [s] Sair   [r] Executar aqui"

	detailsPanel := panelStyle.Copy().
		Width(rightPanelWidth).
		Height(mainPanelHeight - 4).
		Render(fileDetailsContent)

	actionsPanel := panelStyle.Copy().
		Width(rightPanelWidth).
		Height(4).
		Render(fileActionsContent)

	rightPanel := lipgloss.JoinVertical(lipgloss.Left, detailsPanel, actionsPanel)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
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
		return borderStyle.Render(lipgloss.NewStyle().Padding(1).Italic(true).Render("Running command..."))
	}
	return ""
}

func (m *model) getFooterContent() string {
	if m.state == stateFileBrowser {
		return "[S] Sair dos Arquivos  [?] Ajuda  [Q] Sair  " + m.footerMsg
	}
	return "[S] Arquivos  [X] Ações  [?] Ajuda  [Q] Sair  " + m.footerMsg
}

func renderList(commands []models.Command, selected int, width int) string {
	var b strings.Builder
	title := titleStyle.Render("Comandos")
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
	b.WriteString(titleStyle.Render("Explorador: "+path) + "\n")
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
		return "Nenhum arquivo selecionado."
	}
	info, err := entry.Info()
	if err != nil {
		return fmt.Sprintf("Erro ao ler informações: %v", err)
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("Detalhes") + "\n")
	details := []string{
		fmt.Sprintf("Nome:  %s", info.Name()),
		fmt.Sprintf("Tamanho: %d bytes", info.Size()),
		fmt.Sprintf("Modo:    %s", info.Mode().String()),
		fmt.Sprintf("Modificado: %s", info.ModTime().Format("2006-01-02 15:04:05")),
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
	{Key: "↑/k, ↓/j", Description: "Navegar na lista"},
	{Key: "x", Description: "Abrir painel de ações"},
	{Key: "s", Description: "Abrir navegador de arquivos"},
	{Key: "←/→", Description: "Navegar nos diretórios (no explorador)"},
	{Key: "r", Description: "Executar comando (no dir. do explorador)"},
	{Key: "q, esc", Description: "Sair do programa"},
	{Key: "?", Description: "Mostrar/ocultar esta ajuda"},
}

func renderHelpView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Ajuda") + "\n")
	for _, h := range helpBindings {
		b.WriteString(fmt.Sprintf("  %-20s %s\n", h.Key, h.Description))
	}
	b.WriteString("\nPressione '?' ou 'Esc' para fechar.")
	return borderStyle.Render(b.String())
}

func renderActionsPanel(actions []string, selected int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Painel de Ações") + "\n")
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
	b.WriteString("\nUse ↑/↓ para navegar, Enter para selecionar.")
	return borderStyle.Render(lipgloss.NewStyle().Padding(1).Render(b.String()))
}
