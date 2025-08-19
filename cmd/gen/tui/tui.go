package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type ErrMsg struct{ Err error }

func (e ErrMsg) Error() string { return e.Err.Error() }

func Run(g func(func(tea.Msg))) (string, bool, error) {
	p := tea.NewProgram(NewModel())

	go g(p.Send)

	m, err := p.Run()
	if err != nil {
		return "", false, err
	}

	finalModel, ok := m.(Model)
	if !ok {
		return "", false, fmt.Errorf("could not cast model to tui.Model")
	}

	return finalModel.Command, finalModel.Confirmed, finalModel.Err
}
