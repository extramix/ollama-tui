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
		Stream: false,
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

func sendToOllama(prompt string, model string) tea.Cmd {
	return func() tea.Msg {
		response, err := NewOllamaClient(model).SendPrompt(prompt)
		return ollamaResponseMsg{response: response, err: err}
	}
}
