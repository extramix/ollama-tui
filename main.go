package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	input       textinput.Model
	messages    []string
	quitting    bool
	waiting     bool
	ollamaModel string
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
	m.ollamaModel = "llama3.2"
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case ollamaResponseMsg:
		m.waiting = false
		if msg.err != nil {
			m.messages = append(m.messages, "Assistant: Error - "+msg.err.Error())
		} else {
			m.messages = append(m.messages, "Assistant: "+msg.response)
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			userInput := m.input.Value()
			m.messages = append(m.messages, "User: "+userInput)
			m.input.Reset()
			m.waiting = true
			return m, sendToOllama(userInput, m.ollamaModel)
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

	termWidth := 80
	s := ""
	for _, msg := range m.messages {
		wrapped := wrapText(msg, termWidth)
		for _, line := range wrapped {
			s += line + "\n"
		}
		s += "\n"
	}

	s += "\n"
	s += fmt.Sprintf("Your input: %s", m.input.View())
	s += "\n"
	s += "\n--------------------\n"
	return s
}

func main() {
	p := tea.NewProgram(&model{})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}
	words := strings.Fields(text)
	wrapped := []string{}
	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+len(word)+1 > width {
			wrapped = append(wrapped, currentLine)
			currentLine = word
		} else {
			currentLine += " " + word
		}
	}
	wrapped = append(wrapped, currentLine)
	return wrapped
}
