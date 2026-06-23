package ai

import "fmt"

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
	return generateViaResponses("OpenRouter", openrouterAPIURL, o.apiKey, o.model, diff,
		map[string]string{"X-Title": "sage"})
}
