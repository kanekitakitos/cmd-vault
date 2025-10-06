package tui

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/you/cmd-vault/internal/db"
	"github.com/you/cmd-vault/internal/models"
)

type state int

const (
	stateNormal state = iota
	stateAdd
	stateEdit
	stateConfirmDelete
	stateHelp
	stateRunningCmd
)

type model struct {
	store    *db.Store
	commands []models.Command
	selected int
	width    int
	height   int
	state    state

	// inputs for add/edit
	nameInput textinput.Model
	cmdInput  textinput.Model
	noteInput textinput.Model

	// temp for edit
	editCommand *models.Command

	// message / footer
	footerMsg string
}

func RunTUI(store *db.Store) error {
	m := initialModel(store)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func initialModel(store *db.Store) model {
	name := textinput.New()
	name.Placeholder = "name (unique)"
	name.CharLimit = 64
	name.Width = 30
	// Define o foco inicial
	name.Focus()

	cmdi := textinput.New()
	cmdi.Placeholder = "command (e.g. echo %PATH%)"
	cmdi.CharLimit = 256
	cmdi.Width = 50

	note := textinput.New()
	note.Placeholder = "note (required)"
	note.CharLimit = 512
	note.Width = 60

	m := model{
		store:     store,
		selected:  0,
		state:     stateNormal,
		nameInput: name,
		cmdInput:  cmdi,
		noteInput: note,
	}
	m.reloadCommands()
	return m
}

func (m *model) reloadCommands() {
	commands, err := m.store.GetAllCommands()
	if err != nil {
		m.footerMsg = "DB error: " + err.Error()
		m.commands = nil
		return
	}
	m.commands = commands
	if m.selected >= len(m.commands) {
		m.selected = max(0, len(m.commands)-1)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateNormal:        return m.updateNormal(msg)
		case stateAdd:           return m.updateAdd(msg)
		case stateEdit:          return m.updateEdit(msg)
		case stateConfirmDelete: return m.updateConfirmDelete(msg)
		case stateHelp:          return m.updateHelp(msg)
		case stateRunningCmd:    // ignore keys while running
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// === CORREÇÃO DOS ERROS DE COMPILAÇÃO ESTÁ AQUI: ATUALIZAR INPUTS ===
	if m.state == stateAdd || m.state == stateEdit {
		// Passa a mensagem para cada input e combina os comandos retornados.
		var newCmd tea.Cmd
		
		m.nameInput, newCmd = m.nameInput.Update(msg)
		cmd = tea.Batch(cmd, newCmd)

		m.cmdInput, newCmd = m.cmdInput.Update(msg)
		cmd = tea.Batch(cmd, newCmd)

		m.noteInput, newCmd = m.noteInput.Update(msg)
		cmd = tea.Batch(cmd, newCmd)
	}

	return m, cmd
}

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(m.commands)-1 {
			m.selected++
		}
	case "a", "A":
		m.state = stateAdd
		m.nameInput.SetValue("")
		m.cmdInput.SetValue("")
		m.noteInput.SetValue("")
		m.footerMsg = "Add mode - fill fields and press Enter to save, Esc to cancel"
		return m, m.nameInput.Focus()
	case "e", "E":
		if len(m.commands) == 0 {
			m.footerMsg = "No command to edit"
			return m, nil
		}
		m.state = stateEdit
		c := m.commands[m.selected]
		m.editCommand = &c
		m.nameInput.SetValue(c.Name)
		m.cmdInput.SetValue(c.CommandStr)
		m.noteInput.SetValue(c.Note)
		m.footerMsg = "Edit mode - change fields and press Enter to save, Esc to cancel"
		return m, m.nameInput.Focus()
	case "d", "D":
		if len(m.commands) == 0 {
			m.footerMsg = "No command to delete"
			return m, nil
		}
		m.state = stateConfirmDelete
		m.footerMsg = "Confirm delete? (y)es / (n)o"
	case "r", "R":
		if len(m.commands) == 0 {
			m.footerMsg = "No command to run"
			return m, nil
		}
		m.state = stateRunningCmd
		go m.runSelectedCommand()
		m.footerMsg = "Running command..."
	case "?":
		m.state = stateHelp
	}
	return m, nil
}

func (m *model) handleTab() {
	if m.nameInput.Focused() {
		m.nameInput.Blur()
		m.cmdInput.Focus()
	} else if m.cmdInput.Focused() {
		m.cmdInput.Blur()
		m.noteInput.Focus()
	} else if m.noteInput.Focused() {
		m.noteInput.Blur()
		m.nameInput.Focus()
	}
}

func (m *model) handleVerticalNav(key string) {
	if key == "down" {
		if m.nameInput.Focused() {
			m.nameInput.Blur()
			m.cmdInput.Focus()
		} else if m.cmdInput.Focused() {
			m.cmdInput.Blur()
			m.noteInput.Focus()
		}
	} else if key == "up" {
		if m.noteInput.Focused() {
			m.noteInput.Blur()
			m.cmdInput.Focus()
		} else if m.cmdInput.Focused() {
			m.cmdInput.Blur()
			m.nameInput.Focus()
		}
	}
}

func (m model) updateAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(m.nameInput.Value())
		cmdStr := strings.TrimSpace(m.cmdInput.Value())
		note := strings.TrimSpace(m.noteInput.Value())
		if name == "" || note == "" {
			m.footerMsg = "Name and Note required"
			return m, nil
		}
		existing, err := m.store.GetByName(name)
		if err != nil {
			m.footerMsg = "DB error: " + err.Error()
			return m, nil
		}
		if existing != nil {
			m.footerMsg = "Name already exists"
			return m, nil
		}
		c := &models.Command{
			Name:       name,
			CommandStr: cmdStr,
			Note:       note,
			CreatedAt:  time.Now(),
		}
		if _, err := m.store.InsertCommand(c); err != nil {
			m.footerMsg = "Failed to insert: " + err.Error()
			return m, nil
		}
		m.reloadCommands()
		m.state = stateNormal
		m.footerMsg = "Added."
		m.nameInput.Blur()
	case "esc":
		m.state = stateNormal
		m.footerMsg = "Cancelled add"
		m.nameInput.Blur()
	case "tab":
		m.handleTab()
	case "up", "down":
		m.handleTab()
	}
	return m, nil
}

func (m model) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.editCommand == nil {
			m.footerMsg = "Nothing to edit"
			m.state = stateNormal
			return m, nil
		}
		name := strings.TrimSpace(m.nameInput.Value())
		cmdStr := strings.TrimSpace(m.cmdInput.Value())
		note := strings.TrimSpace(m.noteInput.Value())
		if name == "" || note == "" {
			m.footerMsg = "Name and Note required"
			return m, nil
		}
		if name != m.editCommand.Name {
			existing, err := m.store.GetByName(name)
			if err != nil {
				m.footerMsg = "DB error: " + err.Error()
				return m, nil
			}
			if existing != nil {
				m.footerMsg = "Name already exists"
				return m, nil
			}
		}
		m.editCommand.Name = name
		m.editCommand.CommandStr = cmdStr
		m.editCommand.Note = note
		if err := m.store.UpdateCommand(m.editCommand); err != nil {
			m.footerMsg = "Update failed: " + err.Error()
			return m, nil
		}
		m.reloadCommands()
		m.state = stateNormal
		m.footerMsg = "Saved."
		m.nameInput.Blur()
	case "esc":
		m.state = stateNormal
		m.footerMsg = "Cancelled edit"
		m.nameInput.Blur()
	case "tab":
		m.handleTab()
	case "up", "down":
		m.handleTab()
	}
	return m, nil
}

func (m model) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if len(m.commands) > 0 {
			id := m.commands[m.selected].ID
			if err := m.store.DeleteCommand(id); err != nil {
				m.footerMsg = "Delete failed: " + err.Error()
			} else {
				m.footerMsg = "Deleted."
			}
			m.reloadCommands()
		} else {
			m.footerMsg = "No command to delete"
		}
		m.state = stateNormal
	case "n", "N", "esc":
		m.state = stateNormal
		m.footerMsg = "Delete cancelled"
	}
	return m, nil
}

func (m model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" || msg.String() == "?" {
		m.state = stateNormal
		m.footerMsg = ""
	}
	return m, nil
}

func (m model) View() string {
	w := m.width
	if w <= 0 {
		w = 100
	}
	leftW := int(float32(w) * 0.35)
	rightW := w - leftW - 4

	var left string
	left = renderList(m.commands, m.selected, leftW)
	left = borderStyle.Render(left)

	var rightTop string
	var rightBottom string
	if len(m.commands) > 0 {
		c := &m.commands[m.selected]
		rightTop = borderStyle.Render(renderDetails(c))
		rightBottom = borderStyle.Render(renderNote(c, rightW))
	} else {
		rightTop = borderStyle.Render("No commands")
		rightBottom = borderStyle.Render("No notes")
	}

	// assemble view
	var b strings.Builder
	// top line with panels
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(leftW).Render(left),
		lipgloss.NewStyle().Width(rightW).Render(rightTop),
	))
	b.WriteString("\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(leftW).Render(""),
		lipgloss.NewStyle().Width(rightW).Render(rightBottom),
	))
	// footer
	footer := footerStyle.Render("[A] Add  [E] Edit  [R] Run  [D] Delete  [?] Help  [Q] Quit  " + m.footerMsg)
	b.WriteString("\n\n")
	b.WriteString(footer)
	// help overlay
	if m.state == stateHelp {
		b.WriteString("\n\n" + renderHelpView())
	}
	// confirm delete overlay
	if m.state == stateConfirmDelete {
		b.WriteString("\n\n" + lipgloss.NewStyle().Bold(true).Render("Confirm delete? (y/n)"))
	}
	// add/edit overlay
	if m.state == stateAdd || m.state == stateEdit {
		title := "Add Command"
		if m.state == stateEdit {
			title = "Edit Command"
		}
		b.WriteString("\n\n" + lipgloss.NewStyle().Bold(true).Render(title))
		b.WriteString("\nName: " + m.nameInput.View())
		b.WriteString("\nCmd:  " + m.cmdInput.View())
		b.WriteString("\nNote: " + m.noteInput.View())
		b.WriteString("\n\nPress Enter to save, Esc to cancel")
	}
	if m.state == stateRunningCmd {
		b.WriteString("\n\n" + lipgloss.NewStyle().Italic(true).Render("Running command... (TUI is paused until command completes)"))
	}
	return b.String()
}

// helper
func max(a, b int) int {
	if a > b {
		return a
	}
	return b //
}

// run command in external shell (cmd /C), blocking and returning to TUI after completion.
func (m *model) runSelectedCommand() {
	if len(m.commands) == 0 {
		m.footerMsg = "No command to run"
		m.state = stateNormal
		return
	}
	c := m.commands[m.selected]
	// increment usage BEFORE running to keep record even if fail
	_ = m.store.IncrementUsage(c.ID)

	// Use cmd /C as required
	cmd := exec.Command("cmd", "/C", c.CommandStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Clear screen to allow command to show nicely:
	// Attempt to restore later by reloading and forcing view update - bubbletea will continue when goroutine finishes.
	_ = cmd.Run()

	// reload after run
	m.reloadCommands()
	m.state = stateNormal
	m.footerMsg = "Command finished"
}