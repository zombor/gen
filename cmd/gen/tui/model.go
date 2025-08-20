package tui

import (
	"context"
	"log/slog"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zombor/gen/llm"
)

type state int

const (
	promptState state = iota
	commandState
)

type Model struct {
	spinner     spinner.Model
	loading     bool
	command     string
	textarea    textarea.Model
	accepted    bool
	prompt      string
	llmProvider llm.LLMProvider
	state       state
}

func NewModel(prompt string, llmProvider llm.LLMProvider) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	ta := textarea.New()
	ta.Placeholder = "Enter your prompt here..."
	ta.Focus()

	m := Model{
		spinner:     s,
		loading:     true,
		prompt:      prompt,
		llmProvider: llmProvider,
		textarea:    ta,
	}

	if prompt == "" {
		m.state = promptState
		m.loading = false
	} else {
		m.state = commandState
	}

	return m
}

func (m Model) Init() tea.Cmd {
	if m.state == commandState {
		return tea.Batch(m.spinner.Tick, m.generateCommand)
	}
	return nil
}

type commandGeneratedMsg struct {
	command string
}

func (m Model) generateCommand() tea.Msg {
	command, err := m.llmProvider.GenerateCommand(context.Background(), slog.Default(), m.prompt, "bash")
	if err != nil {
		return tea.Quit
	}
	return commandGeneratedMsg{command: command}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.textarea.SetWidth(msg.Width)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+s":
			if m.state == promptState {
				m.prompt = m.textarea.Value()
				m.state = commandState
				m.loading = true
				m.textarea.Reset()
				m.textarea.Placeholder = "Enter your command here..."
				return m, tea.Batch(m.spinner.Tick, m.generateCommand)
			}
			m.accepted = true
			return m, tea.Quit
		}
	case commandGeneratedMsg:
		m.loading = false
		m.command = msg.command
		m.textarea.SetValue(msg.command)
	}

	m.spinner, cmd = m.spinner.Update(msg)
	m.textarea, _ = m.textarea.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	if m.loading {
		return m.spinner.View() + " Thinking..."
	}

	if m.state == promptState {
		return "Enter a prompt to generate a command:\n\n" + m.textarea.View() + "\n\n(ctrl+s to submit, ctrl+c to quit)"
	}

	return m.textarea.View() + "\n\n(ctrl+s to accept, ctrl+c to quit)"
}

func (m Model) Accepted() bool {
	return m.accepted
}

func (m Model) Command() string {
	return m.textarea.Value()
}
