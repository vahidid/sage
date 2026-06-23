package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	freeLLMAPIDefaultModel = "auto"
)

// FreeLLMAPIProvider calls the FreeLLMApi OpenAI-compatible API.
type FreeLLMAPIProvider struct {
	apiKey  string
	baseURL string
	model   string
}

func NewFreeLLMAPIProvider(apiKey, baseURL, model string) *FreeLLMAPIProvider {
	if model == "" {
		model = freeLLMAPIDefaultModel
	}
	return &FreeLLMAPIProvider{apiKey: apiKey, baseURL: baseURL, model: model}
}

func (f *FreeLLMAPIProvider) Name() string {
	return fmt.Sprintf("freellmapi (%s)", f.model)
}

func (f *FreeLLMAPIProvider) GenerateCommitMessage(diff string) (string, error) {
	body, err := json.Marshal(newOpenAICompatibleResponsesRequest(f.model, freeLLMAPICommitMessages(diff)))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, freeLLMAPIResponsesURL(f.baseURL), bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var fr openAICompatibleResponsesResponse
	if err := json.Unmarshal(raw, &fr); err != nil {
		return "", formatProviderParseError("FreeLLMApi", resp.StatusCode, raw, err)
	}

	if resp.StatusCode >= http.StatusBadRequest && fr.Error == nil {
		return "", formatProviderAPIError("FreeLLMApi", resp.StatusCode, providerAPIError{
			Message: resp.Status,
		}, raw)
	}

	if fr.Error != nil {
		return "", formatProviderAPIError("FreeLLMApi", resp.StatusCode, *fr.Error, raw)
	}

	content := firstResponsesText(fr)
	if content == "" {
		return "", formatProviderEmptyResponse("FreeLLMApi", resp.StatusCode, raw)
	}

	message, err := parseStructuredCommitMessage(content)
	if err != nil {
		return "", fmt.Errorf("FreeLLMApi returned invalid structured output: %w", err)
	}
	return message, nil
}

func freeLLMAPIResponsesURL(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(baseURL, "/responses") {
		return baseURL
	}
	if strings.HasSuffix(baseURL, "/v1") {
		return baseURL + "/responses"
	}
	return baseURL + "/v1/responses"
}
