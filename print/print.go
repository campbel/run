package print

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var colors = []string{
	"#89d6fb",
	"#02a9f7",
	"#02577a",
	"#6522a3",
	"#52096a",
}

var index = 0

func StartInfoContext() func(string, ...any) {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color(colors[index]))
	index = (index + 1) % len(colors)
	return func(format string, a ...any) {
		fmt.Println(style.Render(fmt.Sprintf(format, a...)))
	}
}
