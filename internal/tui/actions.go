package tui

import (
	"bytes"
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
			return cmdFinishedMsg{err: nil, output: []byte("No command to run.")} // No command to run, just finish
		}
		c := m.commands[m.selected]
		_ = m.store.IncrementUsage(c.ID)

		var out bytes.Buffer
		cmd := exec.Command("cmd", "/C", c.CommandStr) // #nosec G204
		cmd.Dir = m.currentPath
		cmd.Stdout = &out
		cmd.Stderr = &out
		cmd.Stdin = os.Stdin
		err := cmd.Run()

		return cmdFinishedMsg{err: err, output: out.Bytes()}
	}
}

// runCustomCommand executes a given command string in the current path.
func (m *model) runCustomCommand(commandStr string) tea.Cmd {
	return func() tea.Msg {
		var out bytes.Buffer
		cmd := exec.Command("cmd", "/C", commandStr) // #nosec G204
		cmd.Dir = m.currentPath
		cmd.Stdout = &out
		cmd.Stderr = &out
		cmd.Stdin = os.Stdin
		err := cmd.Run()

		// We don't increment usage as this is a one-off command
		return cmdFinishedMsg{err: err, output: out.Bytes()}
	}
}
