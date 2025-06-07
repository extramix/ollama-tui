# 🦙 Ollama TUI

A beautiful terminal user interface for conversing with local AI models through Ollama. Built with Go and Bubblegum.

![Ollama TUI Demo](https://img.shields.io/badge/Go-1.24+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

## ✨ Features

- **🎨 Beautiful Terminal Interface**: Modern, colorful TUI with smooth scrolling and responsive design
- **💬 Real-time Streaming**: Watch AI responses appear in real-time as they're generated
- **🖱️ Mouse & Keyboard Support**: Scroll through chat history with mouse wheel or arrow keys
- **📝 Markdown Formatting**: Automatic formatting for headers, bold text, and bullet points
- **🔄 Chat History**: Scroll through previous conversations with full context
- **⚡ Fast & Lightweight**: Built with Go for optimal performance
- **🎯 Smart Text Wrapping**: Intelligent word wrapping for optimal readability

## 🚀 Quick Start

### Prerequisites

1. **Install Ollama**: Download and install [Ollama](https://ollama.ai/) on your system
2. **Pull a Model**: Run `ollama pull llama3.2` (or your preferred model)
3. **Start Ollama Service**: Ensure Ollama is running on `localhost:11434`

### Installation

#### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/ollama-tui.git
cd ollama-tui

# Build the application
go build -o ollama-tui

# Run the application
./ollama-tui
```

#### Option 2: Direct Run

```bash
# Clone and run directly
git clone https://github.com/yourusername/ollama-tui.git
cd ollama-tui
go run .
```

## 🎮 Usage

### Basic Controls

- **Type & Enter**: Send messages to the AI
- **↑/↓ Arrow Keys**: Scroll through chat history
- **Page Up/Down**: Quick scroll through long conversations
- **Mouse Wheel**: Scroll through messages
- **Ctrl+C or ESC**: Exit the application

## ⚙️ Configuration

### Changing the AI Model

By default, the application uses `llama3.2`. To use a different model:

1. Ensure the model is available in Ollama: `ollama list`
2. Modify the `ollamaModel` variable in `main.go`:

```go
m.ollamaModel = "your-preferred-model"
```

3. Rebuild the application

### Customizing the Interface

The application uses [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling. You can customize colors and styles by modifying the style definitions in `main.go`.

## 🏗️ Architecture

### Project Structure

```
ollama-tui/
├── main.go      # Main TUI application and UI logic
├── ollama.go    # Ollama API client and streaming logic
├── go.mod       # Go module dependencies
├── go.sum       # Dependency checksums
└── README.md    # This file
```

### Key Components

- **Bubble Tea Model**: Manages application state and UI updates
- **Viewport**: Handles scrollable chat history
- **Text Input**: Manages user input with validation
- **Spinner**: Shows loading state during AI responses
- **Streaming Client**: Handles real-time response streaming from Ollama

### Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling and layout

## 🛠️ Development

### Building

```bash
# Build for current platform
go build -o ollama-tui

# Build for specific platforms
GOOS=linux GOARCH=amd64 go build -o ollama-tui-linux
GOOS=windows GOARCH=amd64 go build -o ollama-tui-windows.exe
GOOS=darwin GOARCH=amd64 go build -o ollama-tui-macos
```

### Code Structure

The application follows a clean architecture pattern:

- **UI Layer** (`main.go`): Handles user interface, input/output, and view rendering
- **Client Layer** (`ollama.go`): Manages communication with Ollama API
- **State Management**: Uses Bubble Tea's Elm-inspired architecture

### Adding Features

To add new features:

1. Define new message types in the appropriate file
2. Handle the messages in the `Update` method
3. Update the `View` method to reflect UI changes
4. Add any new API calls to `ollama.go`

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Setup

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and test thoroughly
4. Commit your changes: `git commit -am 'Add some feature'`
5. Push to the branch: `git push origin feature-name`
6. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

**Made with ❤️ and Go** 