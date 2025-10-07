package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/textinput"
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
)

// cmdFinishedMsg is sent when a command finishes running.
type cmdFinishedMsg struct {
	err error
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

	wd, err := os.Getwd()
	if err != nil {
		wd = "." // Fallback to relative path on error
	}

	m := model{
		store:          store,
		selected:       0,
		state:          stateNormal,
		nameInput:      name,
		cmdInput:       cmdi,
		noteInput:      note,
		currentPath:    wd,
		actions:        []string{"Adicionar Comando", "Editar Comando", "Executar Comando", "Deletar Comando"},
		selectedAction: 0,
	}

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
		if msg.String() == "q" && m.state != stateAdd && m.state != stateEdit {
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
		case stateRunningCmd: // ignore keys while running
			return m, nil
		}
	case cmdFinishedMsg:
		m.state = stateNormal
		if msg.err != nil {
			m.footerMsg = "Command failed: " + msg.err.Error()
		} else {
			m.footerMsg = "Command finished."
		}
		m.reloadCommands()
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
