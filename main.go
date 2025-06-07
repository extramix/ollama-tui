package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

type model struct {
	input           textinput.Model
	viewport        viewport.Model
	currentResponse string
	streamResponse  *http.Response
	messages        []string
	quitting        bool
	waiting         bool
	ollamaModel     string
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
	m.viewport = viewport.New(80, 10)
	m.viewport.SetContent(m.View())
	m.viewport.GotoBottom()
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case startStreamMsg:
		m.currentResponse = ""
		m.messages = append(m.messages, "Assistant:")
		m.streamResponse = msg.response
		return m, streamResponse(msg.response)
	case streamTokenMsg:
		if msg.err != nil {
			m.waiting = false
			m.messages[len(m.messages)-1] = "Assistant: Error - " + msg.err.Error()
			m.streamResponse = nil
			return m, nil
		}
		if msg.done {
			m.waiting = false
			m.streamResponse = nil
			return m, nil
		}
		m.currentResponse += msg.token
		m.messages[len(m.messages)-1] = "Assistant: " + m.currentResponse
		m.viewport.SetContent(m.View())
		m.viewport.GotoBottom()
		return m, streamResponse(m.streamResponse)
	case ollamaResponseMsg:
		m.waiting = false
		if msg.err != nil {
			m.messages = append(m.messages, "Assistant: Error - "+msg.err.Error())
		} else {
			m.messages = append(m.messages, "Assistant: "+msg.response)
		}
		m.viewport.SetContent(m.View())
		m.viewport.GotoBottom()
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

	termWidth, _ := getTerminalSize()
	s := ""
	for _, msg := range m.messages {
		wrapped := wrapText(msg, termWidth-4)
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

func getTerminalSize() (int, int) {
	fd := os.Stdout.Fd()
	width, height, err := term.GetSize(int(fd))
	if err != nil {
		return 80, 20
	}
	return width, height
}
