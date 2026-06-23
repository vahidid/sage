package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// generateViaResponses runs the full commit-message flow against any
// OpenAI-compatible Responses API endpoint. The hosted providers (OpenAI,
// OpenRouter, FreeLLMApi) differ only in name, URL, and extra headers.
func generateViaResponses(provider, url, apiKey, model, diff string, extraHeaders map[string]string) (string, error) {
	messages := structuredCommitMessages(diff)
	var lastInvalid error

	for attempt := 0; attempt < 2; attempt++ {
		body, err := json.Marshal(newOpenAICompatibleResponsesRequest(model, messages))
		if err != nil {
			return "", fmt.Errorf("failed to build request: %w", err)
		}

		debugLogRequest(provider, url, body)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
		for key, value := range extraHeaders {
			req.Header.Set(key, value)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("request failed: %w", err)
		}

		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		debugLogResponse(provider, resp.StatusCode, raw)

		var parsed openAICompatibleResponsesResponse
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return "", formatProviderParseError(provider, resp.StatusCode, raw, err)
		}

		if resp.StatusCode >= http.StatusBadRequest && parsed.Error == nil {
			return "", formatProviderAPIError(provider, resp.StatusCode, providerAPIError{
				Message: resp.Status,
			}, raw)
		}
		if parsed.Error != nil {
			return "", formatProviderAPIError(provider, resp.StatusCode, *parsed.Error, raw)
		}

		content, err := extractResponsesText(provider, parsed, resp.StatusCode, raw)
		if err != nil {
			return "", err
		}

		message, err := parseProviderCommitMessage(content)
		if err == nil {
			return message, nil
		}
		lastInvalid = err
		messages = append(messages, openAICompatibleMessage{
			Role:    "user",
			Content: fmt.Sprintf("Previous output was invalid because: %v. Return only one valid Conventional Commit message for the same diff.", err),
		})
	}

	return "", fmt.Errorf("%s returned invalid structured output: %w", provider, lastInvalid)
}
