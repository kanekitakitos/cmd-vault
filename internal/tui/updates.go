package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kanekitakitos/cmd-vault/internal/models"
)

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(m.commands)-1 {
			m.selected++
		}
	case "tab":
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
		m.previousState = m.state
		m.state = stateRunningCmd
		return m, m.runSelectedCommand()
	case "s", "S":
		m.state = stateFileBrowser
		m.selectedFile = 0
		m.reloadFiles()
		m.footerMsg = "File Browser - [Arrows] to navigate, [s] to exit, [r] to run"
	case "?":
		m.state = stateHelp
	case "x", "X":
		m.previousState = m.state
		m.state = stateActionsPanel
		m.state = stateContextHelp
		m.footerMsg = "Press 'x' again to close help"
	case "o", "O":
		m.previousState = m.state
		m.state = stateOutputFocus
		m.footerMsg = "Output Focus - [↑/↓] to scroll, [Esc] or [o] to exit"
	}
	return m, nil
}

func (m model) updateFileBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedFile > 0 {
			m.selectedFile--
		}
	case "down", "j":
		if m.selectedFile < len(m.files)-1 {
			m.selectedFile++
		}
	case "right", "enter":
		if len(m.files) > 0 {
			selectedEntry := m.files[m.selectedFile]
			if selectedEntry.IsDir() {
				m.currentPath = filepath.Join(m.currentPath, selectedEntry.Name())
				m.selectedFile = 0
				m.reloadFiles()
			}
		}
	case "left", "backspace":
		parentDir := filepath.Dir(m.currentPath)
		if parentDir != m.currentPath {
			m.currentPath = parentDir
			m.selectedFile = 0
			m.reloadFiles()
		}
	case "s", "S", "esc":
		m.state = stateNormal
		m.footerMsg = ""
	case "r", "R":
		m.previousState = m.state
		m.state = stateRunInPath
		m.runInput.SetValue("")
		m.footerMsg = "Enter command to run in current path"
		return m, m.runInput.Focus()
	case "o", "O":
		m.previousState = m.state
		m.state = stateOutputFocus
		return m, m.runInput.Focus()
	case "x", "X":
		m.previousState = m.state
		m.state = stateContextHelp
		m.footerMsg = "Press 'x' again to close help"
	case "c", "C":
		clipboard.WriteAll(m.currentPath)
		m.footerMsg = fmt.Sprintf("Path copied: %s", m.currentPath)

	}
	return m, nil
}

func (m model) updateActionsPanel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedAction > 0 {
			m.selectedAction--
		}
	case "down", "j":
		if m.selectedAction < len(m.actions)-1 {
			m.selectedAction++
		}
	case "enter":
		selectedAction := m.actions[m.selectedAction]
		m.state = stateNormal
		m.footerMsg = ""
		switch selectedAction {
		case "Add Command":
			return m.updateNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		case "Edit Command":
			return m.updateNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
		case "Delete Command":
			return m.updateNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		}
	case "esc", "x", "q":
		m.state = stateNormal
		m.footerMsg = ""
	}
	return m, nil
}

func (m model) updateRunInPath(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		commandStr := strings.TrimSpace(m.runInput.Value())
		if commandStr == "" {
			m.state = stateFileBrowser
			m.footerMsg = "Run cancelled. No command entered."
			return m, nil
		}
		// Prepend command to output and clear input for next command
		newOutput := m.commandOutput + "\n> " + commandStr
		m.commandOutput = newOutput
		m.outputViewport.SetContent(newOutput)
		m.previousState = m.state
		m.state = stateRunningCmd
		return m, m.runCustomCommand(commandStr)
	case "esc", "alt+q":
		m.state = stateFileBrowser
		m.footerMsg = "File Browser - [Arrows] to navigate, [s] to exit, [r] to run"
		m.runInput.Blur()
		return m, nil
	case "p", "P":
		m.previousState = m.state
		m.state = stateSelectCmdToPaste
		return m, nil
	}

	m.runInput, cmd = m.runInput.Update(msg)
	return m, cmd
}

func (m model) updateOutputFocus(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "esc", "o", "O", "q":
		m.state = m.previousState
		if m.state == stateFileBrowser {
			m.footerMsg = "File Browser - [Arrows] to navigate, [s] to exit, [r] to run"
		} else {
			m.footerMsg = ""
		}
	}
	m.outputViewport, cmd = m.outputViewport.Update(msg)
	return m, cmd
}

func (m model) updateContextHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Any key press exits context help
	m.state = m.previousState
	if m.state == stateFileBrowser {
		m.footerMsg = "File Browser - [Arrows] to navigate, [s] to exit, [r] to run"
	} else {
		m.footerMsg = ""
	}
	return m, nil
}

func (m model) updateSelectCmdToPaste(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(m.commands)-1 {
			m.selected++
		}
	case "enter":
		if len(m.commands) > 0 {
			selectedCmd := m.commands[m.selected]
			currentInput := m.runInput.Value()
			m.runInput.SetValue(currentInput + selectedCmd.CommandStr)
			m.runInput.SetCursor(len(m.runInput.Value()))
		}
		m.state = m.previousState // Go back to stateRunInPath
	case "esc", "q", "p":
		m.state = m.previousState // Go back to stateRunInPath
	}

	return m, nil
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
		m.state = m.previousState
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
