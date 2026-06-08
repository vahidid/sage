package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	openrouterAPIURL       = "https://openrouter.ai/api/v1/chat/completions"
	openrouterDefaultModel = "google/gemma-3-12b-it:free"
)

// OpenRouterProvider calls the OpenRouter API (OpenAI-compatible).
type OpenRouterProvider struct {
	apiKey string
	model  string
}

func NewOpenRouterProvider(apiKey, model string) *OpenRouterProvider {
	if model == "" {
		model = openrouterDefaultModel
	}
	return &OpenRouterProvider{apiKey: apiKey, model: model}
}

func (o *OpenRouterProvider) Name() string {
	return fmt.Sprintf("openrouter (%s)", o.model)
}

func (o *OpenRouterProvider) GenerateCommitMessage(diff string) (string, error) {
	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type reqBody struct {
		Model     string `json:"model"`
		MaxTokens int    `json:"max_tokens"`
		Messages  []msg  `json:"messages"`
	}
	type respBody struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	body, err := json.Marshal(reqBody{
		Model:     o.model,
		MaxTokens: 150,
		Messages:  []msg{{Role: "user", Content: buildPrompt(diff)}},
	})
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, openrouterAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("X-Title", "sage")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var or respBody
	if err := json.Unmarshal(raw, &or); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if or.Error != nil {
		return "", fmt.Errorf("OpenRouter API error: %s", or.Error.Message)
	}

	if len(or.Choices) == 0 {
		return "", fmt.Errorf("empty response from OpenRouter")
	}

	return cleanMessage(or.Choices[0].Message.Content), nil
}
