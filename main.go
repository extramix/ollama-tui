package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	input    textinput.Model
	messages []string
	quitting bool
}

// Init initializes the model. It can return a command.
func (m *model) Init() tea.Cmd {
	ti := textinput.New()
	ti.Placeholder = "What's going on?"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 20
	ti.Prompt = "> "
	m.input = ti
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// Handle user pressing Enter (e.g., send input to Ollama)
			m.messages = append(m.messages, "User: "+m.input.Value())
			// TODO: Send m.input to Ollama here and append response
			m.input.Reset() // Clear input after sending
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		}
	}
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Exiting..."
	}

	s := "Ollama TUI Chat\n\n"
	for _, msg := range m.messages {
		s += msg + "\n"
	}
	s += "\n--------------------\n"
	s += fmt.Sprintf("Your input: %s", m.input.View())
	return s
}

func main() {
	p := tea.NewProgram(&model{})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
