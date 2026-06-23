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
		Prompt: plainPrompt(diff),
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", o.host)
	debugLogRequest("Ollama", url, body)

	resp, err := http.Post(
		url,
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
	debugLogResponse("Ollama", resp.StatusCode, raw)

	var or ollamaResponse
	if err := json.Unmarshal(raw, &or); err != nil {
		return "", formatProviderParseError("Ollama", resp.StatusCode, raw, err)
	}

	if resp.StatusCode >= http.StatusBadRequest && or.Error == "" {
		return "", formatProviderAPIError("Ollama", resp.StatusCode, providerAPIError{
			Message: resp.Status,
		}, raw)
	}

	if or.Error != "" {
		return "", formatProviderAPIError("Ollama", resp.StatusCode, providerAPIError{
			Message: or.Error,
		}, raw)
	}

	return cleanMessage(or.Response), nil
}
