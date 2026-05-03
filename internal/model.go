package internal

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

var frameStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	Padding(0, 1).
	BorderForeground(lipgloss.Color("#ffffff"))

type State int8

const (
	InitState State = iota
	ReadyState
	ProcessingState
)

func clamp(value, min, max int) int {
	if value > max {
		return max
	}
	if value < min {
		return min
	}
	return value
}

type Model struct {
	state         State
	textArea      textarea.Model
	viewPort      viewport.Model
	stateSpinner  spinner.Model
	bedrock       *BedrockClient
	messages      []types.Message
	err           error
	width, height int
}

func NewModel(bedrock *BedrockClient) *Model {
	textArea := textarea.New()
	textArea.ShowLineNumbers = false
	textArea.SetVirtualCursor(true)
	textArea.Prompt = ""
	textArea.Focus()
	taStyles := textArea.Styles()
	taStyles.Focused.Base = frameStyle
	taStyles.Blurred.Base = frameStyle
	taStyles.Focused.CursorLine = lipgloss.NewStyle()
	taStyles.Blurred.CursorLine = lipgloss.NewStyle()
	textArea.SetStyles(taStyles)
	viewPort := viewport.New()
	viewPort.Style = frameStyle
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Margin(1, 1).Foreground(lipgloss.Color("205"))
	return &Model{
		state:        InitState,
		stateSpinner: s,
		viewPort:     viewPort,
		textArea:     textArea,
		bedrock:      bedrock,
	}
}

type ResponseMsg types.Message
type ErrorMsg error

func (m Model) SendMessage() tea.Msg {
	out, err := m.bedrock.RuntimeClient.Converse(context.Background(), &bedrockruntime.ConverseInput{
		ModelId:  aws.String(m.bedrock.Config.ModelId),
		Messages: m.messages,
	})
	if err != nil {
		return ErrorMsg(err)
	}
	o := out.Output.(*types.ConverseOutputMemberMessage)
	return ResponseMsg(o.Value)
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.stateSpinner.Tick, textarea.Blink)
}

func (m Model) renderMessages() string {
	var b strings.Builder
	length := len(m.messages)
	for i, msg := range m.messages {
		switch msg.Role {
		case types.ConversationRoleUser:
			b.WriteString("[Me]: ")
		case types.ConversationRoleAssistant:
			b.WriteString("[LLM]: ")
		}
		for _, content := range msg.Content {
			switch c := content.(type) {
			case *types.ContentBlockMemberText:
				b.WriteString(lipgloss.Wrap(c.Value, m.width-frameStyle.GetHorizontalFrameSize(), " "))
			}
		}
		if i != length-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var commands []tea.Cmd
	m.textArea, cmd = m.textArea.Update(msg)
	commands = append(commands, cmd)
	if m.state == ProcessingState {
		m.stateSpinner, cmd = m.stateSpinner.Update(msg)
	}
	commands = append(commands, cmd)
	switch msg := msg.(type) {
	case ErrorMsg:
		m.err = msg
	case ResponseMsg:
		m.messages = append(m.messages, types.Message(msg))
		m.textArea.Focus()
		m.state = ReadyState
		m.viewPort.SetContent(m.renderMessages())
		m.viewPort.GotoBottom()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			m.messages = append(m.messages, types.Message{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: m.textArea.Value(),
					},
				},
			})
			m.textArea.Reset()
			m.textArea.Blur()
			m.state = ProcessingState
			m.viewPort.SetContent(m.renderMessages())
			m.viewPort.GotoBottom()
			commands = append(commands, m.SendMessage)
		case "esc":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		w := msg.Width
		h := msg.Height
		m.width = w
		m.height = h
		m.textArea.SetWidth(w)
		m.viewPort.SetWidth(w)
		frameHeight := frameStyle.GetVerticalFrameSize()
		taHeight := clamp(int(0.2*float32(h)), 1, 4)
		m.textArea.SetHeight(taHeight)
		vpHeight := h - lipgloss.Height(m.stateSpinner.View()) - taHeight - frameHeight
		m.viewPort.SetHeight(vpHeight)
		m.viewPort.GotoBottom()
		m.state = ReadyState
	}
	return m, tea.Batch(commands...)
}

func (m Model) View() tea.View {
	if m.state == InitState {
		return tea.NewView("Initializing...")
	}
	var statusLine string
	if m.err != nil {
		statusLine = m.err.Error()
	}
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.viewPort.View(),
		lipgloss.JoinHorizontal(lipgloss.Center,
			m.stateSpinner.View(),
			lipgloss.Wrap(statusLine, m.width, "")),
		m.textArea.View(),
	)
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
