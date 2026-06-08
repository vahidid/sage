package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	openaiAPIURL       = "https://api.openai.com/v1/chat/completions"
	openaiDefaultModel = "gpt-4o-mini"
)

// OpenAIProvider calls the OpenAI Chat Completions API.
type OpenAIProvider struct {
	apiKey string
	model  string
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = openaiDefaultModel
	}
	return &OpenAIProvider{apiKey: apiKey, model: model}
}

func (o *OpenAIProvider) Name() string {
	return fmt.Sprintf("openai (%s)", o.model)
}

// ── request / response structs ────────────────────────────────────────────────

type openaiRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
	Stream      bool            `json:"stream"`
	Messages    []openaiMessage `json:"messages"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ── main method ───────────────────────────────────────────────────────────────

func (o *OpenAIProvider) GenerateCommitMessage(diff string) (string, error) {
	body, err := json.Marshal(openaiRequest{
		Model:       o.model,
		MaxTokens:   80,
		Temperature: 0,
		Stream:      false,
		Messages: []openaiMessage{
			{Role: "system", Content: commitSystemPrompt},
			{Role: "user", Content: buildPrompt(diff)},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, openaiAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var or openaiResponse
	if err := json.Unmarshal(raw, &or); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if or.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", or.Error.Message)
	}

	if len(or.Choices) == 0 {
		return "", fmt.Errorf("empty response from OpenAI")
	}

	return cleanMessage(or.Choices[0].Message.Content), nil
}
