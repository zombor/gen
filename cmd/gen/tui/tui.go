package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func Run(m tea.Model) (tea.Model, error) {
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr), tea.WithInput(os.Stdin))
	return p.Run()
}
