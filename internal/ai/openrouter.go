package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	openrouterAPIURL       = "https://openrouter.ai/api/v1/responses"
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
	body, err := json.Marshal(newOpenAICompatibleResponsesRequest(o.model, structuredCommitMessages(diff)))
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

	var or openAICompatibleResponsesResponse
	if err := json.Unmarshal(raw, &or); err != nil {
		return "", formatProviderParseError("OpenRouter", resp.StatusCode, raw, err)
	}

	if resp.StatusCode >= http.StatusBadRequest && or.Error == nil {
		return "", formatProviderAPIError("OpenRouter", resp.StatusCode, providerAPIError{
			Message: resp.Status,
		}, raw)
	}

	if or.Error != nil {
		return "", formatProviderAPIError("OpenRouter", resp.StatusCode, *or.Error, raw)
	}

	content := firstResponsesText(or)
	if content == "" {
		return "", formatProviderEmptyResponse("OpenRouter", resp.StatusCode, raw)
	}

	message, err := parseStructuredCommitMessage(content)
	if err != nil {
		return "", fmt.Errorf("OpenRouter returned invalid structured output: %w", err)
	}
	return message, nil
}
