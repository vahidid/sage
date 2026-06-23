package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	openaiAPIURL       = "https://api.openai.com/v1/responses"
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

// ── main method ───────────────────────────────────────────────────────────────

func (o *OpenAIProvider) GenerateCommitMessage(diff string) (string, error) {
	body, err := json.Marshal(newOpenAICompatibleResponsesRequest(o.model, structuredCommitMessages(diff)))
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

	var or openAICompatibleResponsesResponse
	if err := json.Unmarshal(raw, &or); err != nil {
		return "", formatProviderParseError("OpenAI", resp.StatusCode, raw, err)
	}

	if resp.StatusCode >= http.StatusBadRequest && or.Error == nil {
		return "", formatProviderAPIError("OpenAI", resp.StatusCode, providerAPIError{
			Message: resp.Status,
		}, raw)
	}

	if or.Error != nil {
		return "", formatProviderAPIError("OpenAI", resp.StatusCode, *or.Error, raw)
	}

	content := firstResponsesText(or)
	if content == "" {
		return "", formatProviderEmptyResponse("OpenAI", resp.StatusCode, raw)
	}

	message, err := parseStructuredCommitMessage(content)
	if err != nil {
		return "", fmt.Errorf("OpenAI returned invalid structured output: %w", err)
	}
	return message, nil
}
