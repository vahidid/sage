package ai

import (
	"fmt"
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
	return generateViaResponses("FreeLLMApi", freeLLMAPIResponsesURL(f.baseURL), f.apiKey, f.model, diff, nil)
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
