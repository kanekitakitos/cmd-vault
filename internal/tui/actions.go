package tui

import (
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

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
	files, err := os.ReadDir(m.currentPath)
	if err != nil {
		m.footerMsg = "Error reading files: " + err.Error()
		return
	}
	m.files = files
}

// run command in external shell (cmd /C), blocking and returning to TUI after completion.
func (m *model) runSelectedCommand() tea.Cmd {
	return func() tea.Msg {
		if len(m.commands) == 0 {
			return cmdFinishedMsg{err: nil} // No command to run, just finish
		}
		c := m.commands[m.selected]
		_ = m.store.IncrementUsage(c.ID)

		cmd := exec.Command("cmd", "/C", c.CommandStr) // #nosec G204
		cmd.Dir = m.currentPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()

		return cmdFinishedMsg{err: err}
	}
}
