package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	ollamaDefaultHost  = "http://localhost:11434"
	ollamaDefaultModel = "qwen2.5-coder:1.5b"
)

// OllamaProvider calls a locally-running Ollama instance.
type OllamaProvider struct {
	host  string
	model string
}

func NewOllamaProvider(host, model string) *OllamaProvider {
	if host == "" {
		host = ollamaDefaultHost
	}
	if model == "" {
		model = ollamaDefaultModel
	}
	return &OllamaProvider{host: host, model: model}
}

func (o *OllamaProvider) Name() string {
	return fmt.Sprintf("ollama (%s)", o.model)
}

// ── request / response structs ────────────────────────────────────────────────

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"` // false → wait for full response
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

// ── main method ───────────────────────────────────────────────────────────────

func (o *OllamaProvider) GenerateCommitMessage(diff string) (string, error) {
	body, err := json.Marshal(ollamaRequest{
		Model:  o.model,
		Prompt: buildPrompt(diff),
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/generate", o.host),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", fmt.Errorf(
			"cannot reach Ollama at %s: %w\n"+
				"   Is it running?  ollama serve\n"+
				"   Model installed? ollama pull %s",
			o.host, err, o.model,
		)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var or ollamaResponse
	if err := json.Unmarshal(raw, &or); err != nil {
		return "", fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	if or.Error != "" {
		return "", fmt.Errorf("Ollama error: %s", or.Error)
	}

	return cleanMessage(or.Response), nil
}
