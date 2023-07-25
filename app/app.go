package app

import (
	"fmt"
	"strings"
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
)

type EventMsg struct {
	Duration time.Duration
	Message  string
}

func (r EventMsg) String() string {
	if r.Duration == 0 {
		return dotStyle.Render(strings.Repeat(".", 30))
	}
	return fmt.Sprintf("✓ %s %s", r.Message,
		durationStyle.Render(r.Duration.String()))
}

type Model struct {
	spinner  spinner.Model
	results  []EventMsg
	quitting bool
}

func NewModel() Model {
	const numLastResults = 5
	s := spinner.New()
	s.Style = spinnerStyle
	return Model{
		spinner: s,
		results: make([]EventMsg, numLastResults),
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
		m.results = append(m.results[1:], msg)
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

	if m.quitting {
		s += "Done!"
	} else {
		s += m.spinner.View() + " Running actions..."
	}

	s += "\n\n"

	for _, res := range m.results {
		s += res.String() + "\n"
	}

	if !m.quitting {
		s += helpStyle.Render("Press any key to exit")
	}

	if m.quitting {
		s += "\n"
	}

	return appStyle.Render(s)
}
