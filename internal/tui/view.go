package tui

import (
	"os"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kanekitakitos/cmd-vault/internal/models"
)

func renderList(commands []models.Command, selected int, width int) string {
	var b strings.Builder
	title := titleStyle.Render("Comandos")
	b.WriteString(title)
	b.WriteString("\n\n")

	for i, c := range commands {
		// Style for the selected item
		style := lipgloss.NewStyle()
		prefix := "  " // Non-selected prefix
		if i == selected {
			style = style.Foreground(primaryColor).Bold(true)
			prefix = "→ " // Selected prefix
		}

		// Truncate name if it's too long
		name := c.Name
		usage := fmt.Sprintf("(%d)", c.UsageCount)
		availableWidth := width - len(prefix) - len(usage) - 1 // -1 for space
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
	return fmt.Sprintf("%s\n\n%s", titleStyle.Render(c.Name), c.CommandStr)
}

func renderNote(c *models.Command, width int) string {
	if c == nil {
		return "No note"
	}
	// simple wrap
	return lipgloss.NewStyle().Width(width).Render(c.Note)
}

func renderFileBrowser(files []os.DirEntry, selected int, path string, width int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Explorador: "+path) + "\n\n")

	dirStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99")) // Blue for dirs

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

type helpBinding struct {
	Key         string
	Description string
}

var helpBindings = []helpBinding{
	{Key: "↑/k, ↓/j", Description: "Navegar na lista"},
	{Key: "x", Description: "Abrir painel de ações"},
	{Key: "a", Description: "Adicionar novo comando"},
	{Key: "e", Description: "Editar comando selecionado"},
	{Key: "d", Description: "Deletar comando selecionado"},
	{Key: "s", Description: "Abrir navegador de arquivos"},
	{Key: "r", Description: "Executar comando selecionado"},
	{Key: "←/→", Description: "Navegar nos diretórios (no explorador)"},
	{Key: "q, esc", Description: "Sair do programa"},
	{Key: "?", Description: "Mostrar/ocultar esta ajuda"},
}

func renderHelpView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Ajuda") + "\n\n")

	for _, h := range helpBindings {
		b.WriteString(fmt.Sprintf("  %-20s %s\n", h.Key, h.Description))
	}

	b.WriteString("\nPressione '?' ou 'Esc' para fechar.")
	return borderStyle.Render(b.String())
}

func renderActionsPanel(actions []string, selected int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Painel de Ações") + "\n\n")

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
