package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kanekitakitos/cmd-vault/internal/db"
	"github.com/kanekitakitos/cmd-vault/internal/models"
)

type viewMode int

type state int

const (
	stateNormal state = iota
	stateAdd
	stateEdit
	stateConfirmDelete
	stateHelp
	stateFileBrowser
	stateConfirmCancel
	stateRunningCmd
	stateActionsPanel
	stateRunInPath
	stateOutputFocus
	stateContextHelp
	stateSelectCmdToPaste
)

// cmdFinishedMsg is sent when a command finishes running.
type cmdFinishedMsg struct {
	err    error
	output []byte
}

type model struct {
	store         *db.Store
	viewMode      viewMode
	commands      []models.Command
	selected      int
	width         int
	height        int
	state         state
	previousState state

	// file browser
	files        []os.DirEntry
	selectedFile int
	currentPath  string

	// actions panel
	actions        []string
	selectedAction int

	// inputs for add/edit
	nameInput textinput.Model
	cmdInput  textinput.Model
	noteInput textinput.Model

	// input for one-off run
	runInput textinput.Model

	// temp for edit
	editCommand *models.Command

	// message / footer
	footerMsg        string
	commandOutput    string // This will hold the wrapped output
	rawCommandOutput string // This will hold the original, unwrapped output

	// viewport for scrolling output
	outputViewport viewport.Model
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
	// Set initial focus
	name.Focus()

	cmdi := textinput.New()
	cmdi.Placeholder = "command (e.g. echo %PATH%)"
	cmdi.CharLimit = 256
	cmdi.Width = 50

	note := textinput.New()
	note.Placeholder = "note (required)"
	note.CharLimit = 512
	note.Width = 60

	run := textinput.New()
	run.Placeholder = "command to run in current path..."
	run.CharLimit = 256
	run.Width = 60

	wd, err := os.Getwd()
	if err != nil {
		wd = "." // Fallback to relative path on error
	}

	m := model{
		store:            store,
		selected:         0,
		state:            stateNormal,
		nameInput:        name,
		commandOutput:    "Command output will be shown here.",
		rawCommandOutput: "Command output will be shown here.",
		outputViewport:   viewport.New(80, 20), // Will be resized
		cmdInput:         cmdi,
		noteInput:        note,
		runInput:         run,
		currentPath:      wd,
		actions:          []string{"Add Command", "Edit Command", "Delete Command"},
		selectedAction:   0,
	}
	m.outputViewport.SetContent(m.commandOutput)

	m.reloadCommands()
	m.reloadFiles()
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Regra global: 'q' ou 'ctrl+c' deve sair, exceto nos formulários de edição/adição
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "q" && m.state != stateAdd && m.state != stateEdit && m.state != stateRunInPath && m.state != stateOutputFocus {
			return m, tea.Quit
		}
		switch m.state {
		case stateNormal:
			return m.updateNormal(msg)
		case stateAdd:
			return m.updateAdd(msg)
		case stateEdit:
			return m.updateEdit(msg)
		case stateConfirmDelete:
			return m.updateConfirmDelete(msg)
		case stateConfirmCancel:
			return m.updateConfirmCancel(msg)
		case stateFileBrowser:
			return m.updateFileBrowser(msg)
		case stateHelp:
			return m.updateHelp(msg)
		case stateActionsPanel:
			return m.updateActionsPanel(msg)
		case stateRunInPath:
			return m.updateRunInPath(msg)
		case stateOutputFocus:
			return m.updateOutputFocus(msg)
		case stateContextHelp:
			return m.updateContextHelp(msg)
		case stateSelectCmdToPaste:
			return m.updateSelectCmdToPaste(msg)
		case stateRunningCmd:
			return m, nil
		}
	case cmdFinishedMsg:
		if m.previousState == stateRunInPath {
			m.state = stateRunInPath
		} else {
			m.state = stateNormal
		}
		var outputStr string
		if msg.err != nil {
			m.footerMsg = "" // Remove footer message on failure
			outputStr = string(msg.output) + "\n\nError: " + msg.err.Error()
		} else {
			m.footerMsg = "" // Remove footer message on success
			outputStr = string(msg.output)
		}

		m.rawCommandOutput = outputStr // Store the raw, unwrapped output

		// Wrap the output string to fit the panel width to prevent breaking the layout.
		outputPanelWidth := m.getOutputPanelWidth()
		wrappedOutput := lipgloss.NewStyle().Width(outputPanelWidth).Render(outputStr)

		m.commandOutput = wrappedOutput // Keep the full, wrapped output for now
		m.outputViewport.SetContent(wrappedOutput)
		m.outputViewport.GotoTop() // Scroll to top to see the new output
		m.reloadCommands()
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Re-wrap the output content on window resize
		outputPanelWidth := m.getOutputPanelWidth()
		m.outputViewport.SetContent(lipgloss.NewStyle().Width(outputPanelWidth).Render(m.rawCommandOutput))
	}

	return m, cmd
}

func (m model) View() string {
	// If screen size is not yet available, don't render
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	return m.renderView()
}

// getOutputPanelWidth calculates the width of the output panel based on the current view mode and window size.
// This is used to wrap long lines in the command output correctly.
func (m *model) getOutputPanelWidth() int {
	// These calculations mirror the logic in views.go
	const panelPadding = 2 // 1 padding on each side
	const verticalLayoutBreakpoint = 80

	if m.state == stateFileBrowser || m.state == stateRunInPath || (m.state == stateRunningCmd && m.previousState == stateRunInPath) {
		leftPanelWidth := int(float32(m.width) * 0.35)
		return m.width - 4 - leftPanelWidth - panelPadding
	} else if m.width < verticalLayoutBreakpoint {
		return m.width - 2 - panelPadding
	} else { // Normal Horizontal
		leftPanelWidth := int(float32(m.width-4) * 0.35)
		return (m.width - 4) - leftPanelWidth - panelPadding
	}
}

// helper
func max(a, b int) int {
	if a > b {
		return a
	}
	return b //
}
