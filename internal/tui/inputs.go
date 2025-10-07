package tui

import tea "github.com/charmbracelet/bubbletea"

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
