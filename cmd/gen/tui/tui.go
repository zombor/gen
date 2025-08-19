package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func Run(g func(func(tea.Msg))) (string, bool) {
	p := tea.NewProgram(NewModel())

	go g(p.Send)

	m, err := p.Run()
	if err != nil {
		return "", false
	}

	finalModel, ok := m.(Model)
	if !ok {
		return "", false
	}

	return finalModel.Command, finalModel.Confirmed
}
