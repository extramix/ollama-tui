package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	spinner         spinner.Model
	ready           bool
}

// Init initializes the model. It can return a command.
func (m *model) Init() tea.Cmd {
	ti := textinput.New()
	ti.Placeholder = "What's going on?"
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 80
	ti.Prompt = " > "
	m.input = ti
	m.ollamaModel = "llama3.2"

	s := spinner.New()
	s.Spinner = spinner.Globe
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	m.spinner = s

	// Initialize with welcome messages
	m.messages = []string{
		"🦙 **Welcome to Ollama TUI!**\n\nA beautiful terminal interface for conversing with local AI models.\n\n## Getting Started\n\n- Ask questions and get intelligent responses\n- Have natural conversations with AI\n- Use arrow keys or mouse to scroll through chat history\n– Press Ctrl+C or ESC to exit anytime\n\n## Quick Tips\n\n- Type your questions naturally - no special commands needed\n- The AI will stream responses in real-time\n- Scroll up to review previous conversations\n- Currently using model: **" + m.ollamaModel + "**\n\nReady to chat! What would you like to know?",
	}

	return tea.EnterAltScreen
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			// Initialize viewport with proper dimensions
			m.viewport = viewport.New(msg.Width, msg.Height-3) // Reserve 3 lines for input area
			m.viewport.YPosition = 0
			m.viewport.MouseWheelEnabled = true
			// Set input width to full terminal width minus prompt and padding
			m.input.Width = msg.Width - len(m.input.Prompt) - 6 // Account for lipgloss padding
			m.ready = true
			m.updateViewportContent()
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 3
			// Update input width when window is resized
			m.input.Width = msg.Width - len(m.input.Prompt) - 6 // Account for lipgloss padding
			m.updateViewportContent()
		}

	case startStreamMsg:
		m.currentResponse = ""
		m.messages = append(m.messages, "🦙")
		m.streamResponse = msg.response
		m.updateViewportContent()
		return m, tea.Batch(streamResponse(msg.response), m.spinner.Tick)

	case streamTokenMsg:
		if msg.err != nil {
			m.waiting = false
			m.messages[len(m.messages)-1] = "🦙 Error - " + msg.err.Error()
			m.streamResponse = nil
			m.updateViewportContent()
			return m, nil
		}
		if msg.done {
			m.waiting = false
			m.streamResponse = nil
			return m, nil
		}
		// Format the new token and add it to the response
		formattedToken := formatOllamaResponse(msg.token, m.currentResponse)
		m.currentResponse += formattedToken
		m.messages[len(m.messages)-1] = "🦙 " + m.currentResponse
		m.updateViewportContent()
		return m, streamResponse(m.streamResponse)

	case ollamaResponseMsg:
		m.waiting = false
		if msg.err != nil {
			m.messages = append(m.messages, "🦙 Error - "+msg.err.Error())
		}
		m.updateViewportContent()
		return m, nil

	case tea.KeyMsg:
		if m.waiting {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.Type {
		case tea.KeyEnter:
			if m.input.Value() == "" {
				return m, nil
			}
			userInput := m.input.Value()
			m.messages = append(m.messages, "🧋 "+userInput)
			m.input.Reset()
			m.waiting = true
			m.updateViewportContent()
			return m, tea.Batch(sendToOllama(userInput, m.ollamaModel), m.spinner.Tick)

		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		// Handle viewport scrolling keys
		case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDown:
			if !m.waiting {
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		}

	case tea.MouseMsg:
		// Handle mouse events for viewport scrolling
		if !m.waiting {
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	}

	if m.waiting {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		// Only update input when not waiting
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.quitting {
		quitStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
		return quitStyle.Render("✨ Thanks for using Ollama TUI! Goodbye! ✨")
	}

	if !m.ready {
		loadingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
		return loadingStyle.Render("\n  🚀 Initializing Ollama TUI...")
	}

	// Create main container style with padding
	mainStyle := lipgloss.NewStyle().
		Width(m.viewport.Width).
		Padding(0, 4)

	var view strings.Builder

	// Render the viewport (chat messages)
	view.WriteString(m.viewport.View())
	view.WriteString("\n")

	// Render the input area with colorful styling
	if m.waiting {
		spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
		view.WriteString(spinnerStyle.Render(fmt.Sprintf("%s lemme cook...", m.spinner.View())))
	} else {
		// Style the input prompt
		promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) // Green
		inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))             // White
		scrollInfoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))         // Gray

		scrollInfo := fmt.Sprintf(" [%d/%d]", m.viewport.YOffset, m.viewport.TotalLineCount())
		view.WriteString(promptStyle.Render("🧋 ") + inputStyle.Render(m.input.View()) + scrollInfoStyle.Render(scrollInfo))
	}

	return mainStyle.Render(view.String())
}

func main() {
	p := tea.NewProgram(&model{}, tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func formatOllamaResponse(newToken, existingResponse string) string {
	if strings.HasPrefix(strings.TrimSpace(newToken), "1. ") ||
		strings.HasPrefix(strings.TrimSpace(newToken), "2. ") ||
		strings.HasPrefix(strings.TrimSpace(newToken), "3. ") ||
		strings.HasPrefix(strings.TrimSpace(newToken), "4. ") ||
		strings.HasPrefix(strings.TrimSpace(newToken), "5. ") {
		if len(existingResponse) > 0 && !strings.HasSuffix(existingResponse, "\n") {
			return "\n" + newToken
		}
	}

	if strings.HasPrefix(strings.TrimSpace(newToken), "* ") ||
		strings.HasPrefix(strings.TrimSpace(newToken), "- ") {
		if len(existingResponse) > 0 && !strings.HasSuffix(existingResponse, "\n") {
			return "\n" + newToken
		}
	}

	if strings.Contains(newToken, ". ") && !strings.Contains(newToken, "..") {
		formatted := strings.ReplaceAll(newToken, ". ", ".\n")
		return formatted
	}

	return newToken
}

func wrapText(text string, width int) []string {
	lines := strings.Split(text, "\n")
	var wrapped []string

	for _, line := range lines {
		// Preserve empty lines for spacing
		if strings.TrimSpace(line) == "" {
			wrapped = append(wrapped, "")
			continue
		}

		line = strings.TrimSpace(line)

		// If line is short enough, add it as-is
		if len(line) <= width {
			wrapped = append(wrapped, line)
			continue
		}

		// Word wrap longer lines
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}

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
	}

	return wrapped
}

func (m *model) updateViewportContent() {
	if !m.ready {
		return
	}

	content := m.renderMessages()
	m.viewport.SetContent(content)
	// Only auto-scroll to bottom if we're already at or near the bottom
	if m.viewport.AtBottom() || m.viewport.YOffset >= m.viewport.TotalLineCount()-m.viewport.Height-3 {
		m.viewport.GotoBottom()
	}
}

func (m *model) renderMessages() string {
	var content strings.Builder

	for i, msg := range m.messages {
		// Apply markdown formatting before wrapping
		formattedMsg := formatMarkdown(msg)
		wrapped := wrapText(formattedMsg, m.viewport.Width-6) // Account for padding
		for _, line := range wrapped {
			content.WriteString(line)
			content.WriteString("\n")
		}

		// Add spacing between messages, but not after the last one
		if i < len(m.messages)-1 {
			content.WriteString("\n")
		}
	}

	// Add some extra lines to ensure there's scrollable content
	content.WriteString("\n\n")

	return content.String()
}

func formatMarkdown(text string) string {
	boldStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	bulletStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)

	for {
		start := strings.Index(text, "**")
		if start == -1 {
			break
		}

		end := strings.Index(text[start+2:], "**")
		if end == -1 {
			break
		}

		end += start + 2

		boldText := text[start+2 : end]

		styledText := boldStyle.Render(boldText)
		text = text[:start] + styledText + text[end+2:]
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Style headers
		if strings.HasPrefix(trimmed, "##") {
			headerText := strings.TrimSpace(strings.TrimPrefix(trimmed, "##"))
			lines[i] = headerStyle.Render("✨ " + headerText + " ✨")
		} else if strings.HasPrefix(trimmed, "•") {
			// Style bullet points
			bulletText := strings.TrimSpace(strings.TrimPrefix(trimmed, "•"))
			lines[i] = bulletStyle.Render("  ▶ ") + bulletText
		} else if strings.HasPrefix(trimmed, "🧋") {
			// Style user messages
			userText := strings.TrimPrefix(trimmed, "🧋 ")
			userEmojiStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
			lines[i] = userEmojiStyle.Render("🧋 ") + userStyle.Render(userText)
		}
	}
	text = strings.Join(lines, "\n")

	return text
}
