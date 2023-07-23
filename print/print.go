package print

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var infoStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4"))

var noticeStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FFD580")).
	Background(lipgloss.Color("#7D56F4"))

func Info(format string, a ...any) {
	fmt.Println(infoStyle.Render(fmt.Sprintf(format, a...)))
}

func Notice(format string, a ...any) {
	fmt.Println(noticeStyle.Foreground(lipgloss.Color("#FFD580")).Render(fmt.Sprintf(format, a...)))
}
