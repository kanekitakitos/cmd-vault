package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/you/cmd-vault/internal/models"
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

type helpBinding struct {
	Key         string
	Description string
}

var helpBindings = []helpBinding{
	{Key: "↑/k, ↓/j", Description: "Navegar na lista"},
	{Key: "a", Description: "Adicionar novo comando"},
	{Key: "e", Description: "Editar comando selecionado"},
	{Key: "d", Description: "Deletar comando selecionado"},
	{Key: "r", Description: "Executar comando selecionado"},
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
