package app

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	spinnerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	dotStyle      = helpStyle.Copy().UnsetMargins()
	durationStyle = dotStyle.Copy()
	appStyle      = lipgloss.NewStyle().Margin(1, 2, 0, 2)

	outputFrameStyle = lipgloss.NewStyle().
				Align(lipgloss.Left, lipgloss.Top)

	actionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)
)

type EventType string

var (
	EventTypeActionFinish EventType = "finish"
	EventTypeActionStart  EventType = "start"
	EventTypeOutput       EventType = "output"
)

type EventMsg struct {
	EventType
	Duration time.Duration
	Message  string
}

func (r EventMsg) String() string {
	return fmt.Sprintf("✓ %s %s", r.Message,
		durationStyle.Render(r.Duration.Round(time.Second).String()))
}

type Model struct {
	spinner  spinner.Model
	actions  []string
	output   []EventMsg
	quitting bool
	height   int
	width    int
}

func NewModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle
	return Model{
		actions: []string{"starting..."},
		spinner: s,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	case EventMsg:
		switch msg.EventType {
		case EventTypeOutput:
			m.output = append(m.output, msg)
		case EventTypeActionStart:
			m.actions = append([]string{msg.Message}, m.actions...)
		case EventTypeActionFinish:
			m.actions = m.actions[1:]
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	var s string

	s += m.spinner.View() + " Running " + actionStyle.Render(m.actions[0]) + "\n\n"

	output := ""
	for _, out := range m.output {
		output += out.Message
	}

	s += outputFrameStyle.MaxHeight(m.height).Render(output)

	return appStyle.Render(s)
}
