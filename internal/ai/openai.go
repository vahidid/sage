package ai

import "fmt"

const (
	openaiAPIURL       = "https://api.openai.com/v1/responses"
	openaiDefaultModel = "gpt-4o-mini"
)

// OpenAIProvider calls the OpenAI Responses API.
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

func (o *OpenAIProvider) GenerateCommitMessage(diff string) (string, error) {
	return generateViaResponses("OpenAI", openaiAPIURL, o.apiKey, o.model, diff, nil)
}
