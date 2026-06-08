package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	claudeAPIURL       = "https://api.anthropic.com/v1/messages"
	claudeAPIVersion   = "2023-06-01"
	claudeDefaultModel = "claude-haiku-4-5-20251001"
)

// ClaudeProvider calls the Anthropic Messages API.
type ClaudeProvider struct {
	apiKey string
	model  string
}

func NewClaudeProvider(apiKey, model string) *ClaudeProvider {
	if model == "" {
		model = claudeDefaultModel
	}
	return &ClaudeProvider{apiKey: apiKey, model: model}
}

func (c *ClaudeProvider) Name() string {
	return fmt.Sprintf("claude (%s)", c.model)
}

// ── request / response structs ────────────────────────────────────────────────

type claudeRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	System      string    `json:"system"`
	Messages    []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ── main method ───────────────────────────────────────────────────────────────

func (c *ClaudeProvider) GenerateCommitMessage(diff string) (string, error) {
	body, err := json.Marshal(claudeRequest{
		Model:       c.model,
		MaxTokens:   80,
		Temperature: 0,
		System:      commitSystemPrompt,
		Messages: []message{
			{Role: "user", Content: buildPrompt(diff)},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, claudeAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", claudeAPIVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var cr claudeResponse
	if err := json.Unmarshal(raw, &cr); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if cr.Error != nil {
		return "", fmt.Errorf("Claude API error: %s", cr.Error.Message)
	}

	for _, block := range cr.Content {
		if block.Type == "text" {
			return cleanMessage(block.Text), nil
		}
	}
	return "", fmt.Errorf("empty response from Claude")
}
