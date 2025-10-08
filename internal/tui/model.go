package tui

import (
	"os"

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
	footerMsg     string
	commandOutput string

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
		store:          store,
		selected:       0,
		state:          stateNormal,
		nameInput:      name,
		commandOutput:  "Command output will be shown here.",
		outputViewport: viewport.New(80, 20), // Will be resized
		cmdInput:       cmdi,
		noteInput:      note,
		runInput:       run,
		currentPath:    wd,
		actions:        []string{"Add Command", "Edit Command", "Delete Command"},
		selectedAction: 0,
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
			m.footerMsg = "Command failed. See output panel."
			outputStr = string(msg.output) + "\n\nError: " + msg.err.Error()
		} else {
			m.footerMsg = "Command finished successfully."
			outputStr = string(msg.output)
		}
		m.commandOutput = outputStr // Keep the full output
		m.outputViewport.SetContent(outputStr)
		m.reloadCommands()
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.outputViewport.Width = m.width // Will be resized properly in view
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

// helper
func max(a, b int) int {
	if a > b {
		return a
	}
	return b //
}
