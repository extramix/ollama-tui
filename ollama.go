package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	tea "github.com/charmbracelet/bubbletea"
)

type OllamaClient struct {
	baseURL string
	model   string
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type ollamaResponseMsg struct {
	response string
	err      error
}

type startStreamMsg struct {
	response *http.Response
}

type streamTokenMsg struct {
	token string
	done  bool
	err   error
}

func NewOllamaClient(model string) *OllamaClient {
	return &OllamaClient{
		baseURL: "http://localhost:11434",
		model:   model,
	}
}

func (c *OllamaClient) SendPrompt(prompt string) (string, error) {
	request := OllamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: true,
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	response, err := http.Post(c.baseURL+"/api/generate", "application/json", bytes.NewBuffer(jsonRequest))
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var ollamaResponse OllamaResponse
	err = json.NewDecoder(response.Body).Decode(&ollamaResponse)
	if err != nil {
		return "", err
	}

	return ollamaResponse.Response, nil
}

func (c *OllamaClient) SendPromptStream(prompt string) tea.Cmd {
	return func() tea.Msg {
		request := OllamaRequest{
			Model:  c.model,
			Prompt: prompt,
			Stream: true,
		}

		jsonRequest, err := json.Marshal(request)
		if err != nil {
			return ollamaResponseMsg{err: err}
		}

		response, err := http.Post(c.baseURL+"/api/generate", "application/json", bytes.NewBuffer(jsonRequest))
		if err != nil {
			return ollamaResponseMsg{err: err}
		}

		return startStreamMsg{response: response}
	}
}

func streamResponse(response *http.Response) tea.Cmd {
	return func() tea.Msg {

		decoder := json.NewDecoder(response.Body)
		var ollamaResponse OllamaResponse
		if err := decoder.Decode(&ollamaResponse); err != nil {
			return ollamaResponseMsg{err: err}
		}
		if ollamaResponse.Done {
			response.Body.Close()
			return ollamaResponseMsg{response: ollamaResponse.Response, err: nil}
		}
		return streamTokenMsg{token: ollamaResponse.Response, done: false, err: nil}
	}
}

func sendToOllama(prompt string, model string) tea.Cmd {
	return NewOllamaClient(model).SendPromptStream(prompt)
}
