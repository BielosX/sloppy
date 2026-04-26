package internal

import (
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
)

type State int8

const (
	InitState State = iota
	ReadyState
)

type Model struct {
	state         State
	textArea      textarea.Model
	width, height int
}

func NewModel() *Model {
	textArea := textarea.New()
	textArea.ShowLineNumbers = false
	textArea.SetVirtualCursor(false)
	textArea.Focus()
	textArea.SetWidth(30)
	textArea.SetHeight(3)
	return &Model{
		state:    InitState,
		textArea: textArea,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			m.textArea, cmd = m.textArea.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.state = ReadyState
	}
	return m, nil
}

func (m Model) View() tea.View {
	if m.state == InitState {
		return tea.NewView("Initializing...")
	}
	return tea.NewView(m.textArea.View())
}
