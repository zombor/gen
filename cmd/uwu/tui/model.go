package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

type Model struct {
	Spinner     spinner.Model
	Loading     bool
	Command     string
	Confirmed   bool
	Quitting    bool
	Renderer    *glamour.TermRenderer
	TextInput   textinput.Model
}

func NewModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

ti := textinput.New()
ti.Focus()

r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(0),
	)
	return Model{Spinner: s, Loading: true, Renderer: r, TextInput: ti}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.Spinner.Tick, textinput.Blink)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.Quitting = true
			return m, tea.Quit
		case "enter":
			if !m.Loading {
				m.Confirmed = true
				m.Command = m.TextInput.Value()
				return m, tea.Quit
			}
		}
	case CommandGeneratedMsg:
		m.Loading = false
		m.Command = string(msg)
		m.TextInput.SetValue(m.Command)
		return m, nil
	}

	m.Spinner, cmd = m.Spinner.Update(msg)
	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.Quitting {
		return ""
	}
	if m.Loading {
		return m.Spinner.View() + " Generating command..."
	}

	return "Generated command:\n" + m.TextInput.View() + "\n\n(enter to confirm, ctrl+c to quit)"
}

type CommandGeneratedMsg string
