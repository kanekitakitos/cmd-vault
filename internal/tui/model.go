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

type viewMode int

const (
	viewCommands viewMode = iota
	viewTemplates
)

type state int

const (
	stateNormal state = iota
	stateAdd
	stateEdit
	stateConfirmDelete
	stateHelp
	stateConfirmCancel
	stateRunningCmd
)

type model struct {
	store    *db.Store
	viewMode viewMode
	commands []models.Command
	selected int
	width    int
	height   int
	state    state
	previousState state

	// file browser
	files []os.DirEntry

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
	m.reloadFiles()
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

func (m *model) reloadFiles() {
	files, err := os.ReadDir(".")
	if err != nil {
		m.footerMsg = "Error reading files: " + err.Error()
		return
	}
	m.files = files
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Regra global: 'q' ou 'ctrl+c' deve sair, exceto nos formulários de edição/adição
		// onde 'esc' é usado para cancelar.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "q" && m.state != stateAdd && m.state != stateEdit {
			return m, tea.Quit
		}
		switch m.state {
		case stateNormal:        return m.updateNormal(msg)
		case stateAdd:           return m.updateAdd(msg)
		case stateEdit:          return m.updateEdit(msg)
		case stateConfirmDelete: return m.updateConfirmDelete(msg)
		case stateConfirmCancel: return m.updateConfirmCancel(msg)
		case stateHelp:          return m.updateHelp(msg)
		case stateRunningCmd:    // ignore keys while running
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update inputs when in add/edit state
	if m.state == stateAdd || m.state == stateEdit {
		var cmds []tea.Cmd
		var newCmd tea.Cmd

		m.nameInput, newCmd = m.nameInput.Update(msg)
		cmds = append(cmds, newCmd)

		m.cmdInput, newCmd = m.cmdInput.Update(msg)
		cmds = append(cmds, newCmd)

		m.noteInput, newCmd = m.noteInput.Update(msg)
		cmds = append(cmds, newCmd)
		cmd = tea.Batch(cmds...)
	}
	return m, cmd
}

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		// TODO: This will need to be adapted when templates are fully implemented
		if m.selected < len(m.commands)-1 {
			m.selected++
		}
	case "tab":
		// Switch between Commands and Templates view
		m.viewMode = (m.viewMode + 1) % 2 // Toggles between 0 and 1
		m.selected = 0
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
	case "q":
		m.previousState = m.state
		m.state = stateConfirmCancel
		m.footerMsg = "Discard changes? (y/n)"
	case "up", "down":
		m.handleVerticalNav(msg.String())
	}
	var newCmd tea.Cmd
	m, newCmd = m.updateInputs(msg)
	return m, newCmd
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
	case "q":
		m.previousState = m.state
		m.state = stateConfirmCancel
		m.footerMsg = "Discard changes? (y/n)"
	case "up", "down":
		m.handleVerticalNav(msg.String())
	}
	var newCmd tea.Cmd
	m, newCmd = m.updateInputs(msg)
	return m, newCmd
}

// updateInputs propagates the message to the focused text input.
func (m model) updateInputs(msg tea.Msg) (model, tea.Cmd) {
	var cmds []tea.Cmd
	var newCmd tea.Cmd

	m.nameInput, newCmd = m.nameInput.Update(msg)
	cmds = append(cmds, newCmd)

	m.cmdInput, newCmd = m.cmdInput.Update(msg)
	cmds = append(cmds, newCmd)

	m.noteInput, newCmd = m.noteInput.Update(msg)
	cmds = append(cmds, newCmd)

	return m, tea.Batch(cmds...)
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
	case "n", "N", "esc", "q":
		m.state = stateNormal
		m.footerMsg = "Delete cancelled"
	}
	return m, nil
}

func (m model) updateConfirmCancel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		m.state = stateNormal
		m.footerMsg = "Cancelled"
		m.nameInput.Blur()
		m.cmdInput.Blur()
		m.noteInput.Blur()
	case "n", "N", "esc", "q":
		m.state = m.previousState // Volta para stateAdd ou stateEdit
		m.footerMsg = "Continuing..."
	}
	return m, nil
}

func (m model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" || msg.String() == "?" || msg.String() == "q" {
		m.state = stateNormal
		m.footerMsg = ""
	}
	return m, nil
}

func (m model) View() string {
	// If screen size is not yet available, don't render
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Define panel dimensions
	// We subtract heights for footer and margins
	mainPanelHeight := m.height - 4
	leftPanelWidth := int(float32(m.width) * 0.35)
	rightPanelWidth := m.width - leftPanelWidth

	// --- Left Panel (List) ---
	leftPanel := panelStyle.Copy().
		Width(leftPanelWidth).
		Height(mainPanelHeight).
		Render(renderList(m.commands, m.selected, leftPanelWidth-2)) // -2 for padding

	// --- Right Panels (Details + Note) ---
	var rightTopContent, rightBottomContent string
	if len(m.commands) > 0 {
		c := &m.commands[m.selected]
		rightTopContent = renderDetails(c)
		rightBottomContent = renderNote(c, rightPanelWidth-2) // -2 for padding
	} else {
		rightTopContent = "No commands"
		rightBottomContent = "No notes"
	}

	rightTopPanel := panelStyle.Copy().
		Width(rightPanelWidth).
		Height(4). // Fixed height for details
		Render(rightTopContent)

	rightBottomPanel := panelStyle.Copy().
		Width(rightPanelWidth).
		Height(mainPanelHeight - 4). // Remaining height
		Render(rightBottomContent)

	rightPanel := lipgloss.JoinVertical(lipgloss.Left, rightTopPanel, rightBottomPanel)

	// --- Main View ---
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// footer
	footer := footerStyle.Render("[A] Add  [E] Edit  [R] Run  [D] Delete  [?] Help  [Q] Quit  " + m.footerMsg)

	// Overlays (Help, Confirm, Add/Edit)
	var overlay string
	// help overlay
	if m.state == stateHelp {
		overlay = renderHelpView()
	}
	// confirm delete overlay
	if m.state == stateConfirmDelete {
		overlay = borderStyle.Render(lipgloss.NewStyle().Padding(1).SetString("Confirm delete? (y/n)").String())
	}
	// confirm cancel add/edit overlay
	if m.state == stateConfirmCancel {
		overlay = borderStyle.Render(lipgloss.NewStyle().Padding(1).SetString("Discard changes? (y/n)").String())
	}
	// add/edit overlay
	if m.state == stateAdd || m.state == stateEdit {
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
		overlay = borderStyle.Render(lipgloss.NewStyle().Padding(1).Render(form))
	}
	if m.state == stateRunningCmd {
		overlay = borderStyle.Render(lipgloss.NewStyle().Padding(1).Italic(true).Render("Running command..."))
	}

	// Final assembly
	mainContent := lipgloss.JoinVertical(lipgloss.Left, mainView, footer)

	// If there's an overlay, place it on top of the main content.
	if overlay != "" {
		// Create a style for the overlay container that will render the main content as its background
		overlayStyle := lipgloss.NewStyle().Width(m.width).Height(m.height)
		// Place the overlay content in the center
		centeredOverlay := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
		return overlayStyle.Render(centeredOverlay)
	}
	return mainContent
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