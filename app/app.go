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

	actionHeaderStyle = lipgloss.NewStyle().
				Width(25).
				Foreground(lipgloss.Color("63"))

	outputHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("63"))

	actionsFrameStyle = lipgloss.NewStyle().
				Width(25).
				Height(5).
				Align(lipgloss.Left, lipgloss.Top)

	outputFrameStyle = lipgloss.NewStyle().
				Height(5).
				Align(lipgloss.Left, lipgloss.Top)
)

type EventType string

var (
	EventTypeActionFinish EventType = "finish"
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
	results  []EventMsg
	output   []EventMsg
	quitting bool
}

func NewModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle
	return Model{
		spinner: s,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	case EventMsg:
		switch msg.EventType {
		case EventTypeOutput:
			m.output = append(m.output, msg)
			if len(m.output) > 5 {
				m.output = m.output[1:]
			}
		default:
			m.results = append(m.results, msg)
			if len(m.results) > 5 {
				m.results = m.results[1:]
			}
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m Model) View() string {
	var s string

	s += m.spinner.View() + " Running actions...\n\n"

	s += lipgloss.JoinHorizontal(lipgloss.Top,
		actionHeaderStyle.Render("ACTIONS"),
		outputHeaderStyle.Render("OUTPUT"),
	)
	s += "\n"

	// Only show the last 5 results
	actions := ""
	for _, res := range m.results {
		actions += res.String() + "\n"
	}

	output := ""
	for _, out := range m.output {
		output += out.Message
	}

	s += lipgloss.JoinHorizontal(lipgloss.Top,
		actionsFrameStyle.Render(actions),
		outputFrameStyle.Render(output),
	)

	if m.quitting {
		s += "\n"
	}
	s += helpStyle.Render("Press any key to exit")

	return appStyle.Render(s)
}
