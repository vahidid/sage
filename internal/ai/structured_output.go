package ai

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const structuredCommitSystemPrompt = "Return a JSON object that matches the provided schema. The commit_message value must be one Conventional Commit message only. No analysis, no reasoning, no markdown, no quotes around the commit message text."

var conventionalCommitLine = regexp.MustCompile(`(?m)\b(feat|fix|refactor|chore|docs|style|test|perf)(\([^)]+\))?: [^\r\n]+`)

type openAICompatibleMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAICompatibleResponsesRequest struct {
	Model           string                    `json:"model"`
	Input           []openAICompatibleMessage `json:"input"`
	Text            openAICompatibleText      `json:"text"`
	MaxOutputTokens int                       `json:"max_output_tokens"`
	Temperature     float64                   `json:"temperature"`
}

type openAICompatibleText struct {
	Format openAICompatibleTextFormat `json:"format"`
}

type openAICompatibleTextFormat struct {
	Type   string                       `json:"type"`
	Name   string                       `json:"name"`
	Schema openAICompatibleSchemaObject `json:"schema"`
	Strict bool                         `json:"strict"`
}

type openAICompatibleSchemaObject struct {
	Type                 string                                  `json:"type"`
	Properties           map[string]openAICompatibleSchemaObject `json:"properties,omitempty"`
	Required             []string                                `json:"required,omitempty"`
	AdditionalProperties *bool                                   `json:"additionalProperties,omitempty"`
	Description          string                                  `json:"description,omitempty"`
}

type structuredCommitMessage struct {
	CommitMessage string `json:"commit_message"`
}

type openAICompatibleResponsesResponse struct {
	OutputText string `json:"output_text,omitempty"`
	Output     []struct {
		Content []struct {
			Text string `json:"text,omitempty"`
		} `json:"content,omitempty"`
	} `json:"output,omitempty"`
	Error *providerAPIError `json:"error,omitempty"`
}

func commitMessageText() openAICompatibleText {
	return openAICompatibleText{
		Format: openAICompatibleTextFormat{
			Type:   "json_schema",
			Name:   "commit_message_response",
			Strict: true,
			Schema: openAICompatibleSchemaObject{
				Type: "object",
				Properties: map[string]openAICompatibleSchemaObject{
					"commit_message": {
						Type:        "string",
						Description: "A single Conventional Commit message, without markdown or explanations.",
					},
				},
				Required:             []string{"commit_message"},
				AdditionalProperties: boolPtr(false),
			},
		},
	}
}

func newOpenAICompatibleResponsesRequest(model string, input []openAICompatibleMessage) openAICompatibleResponsesRequest {
	return openAICompatibleResponsesRequest{
		Model:           model,
		Input:           input,
		Text:            commitMessageText(),
		MaxOutputTokens: 80,
		Temperature:     0,
	}
}

func boolPtr(v bool) *bool {
	return &v
}

func structuredCommitMessages(diff string) []openAICompatibleMessage {
	return []openAICompatibleMessage{
		{Role: "system", Content: structuredCommitSystemPrompt},
		{Role: "user", Content: buildPrompt(diff)},
	}
}

func freeLLMAPICommitMessages(diff string) []openAICompatibleMessage {
	return []openAICompatibleMessage{
		{Role: "system", Content: structuredCommitSystemPrompt},
		{Role: "user", Content: buildPrompt(diff)},
		{Role: "assistant", Content: "{"},
	}
}

func parseStructuredCommitMessage(content string) (string, error) {
	if message, ok := parseCommitMessageJSON(content); ok {
		return message, nil
	}

	cleaned := stripModelChatter(content)
	if message, ok := parseCommitMessageJSON(cleaned); ok {
		return message, nil
	}
	if !strings.HasPrefix(strings.TrimSpace(cleaned), "{") {
		if message, ok := parseCommitMessageJSON("{" + cleaned); ok {
			return message, nil
		}
	}
	if jsonObject := extractJSONObject(cleaned); jsonObject != "" {
		if message, ok := parseCommitMessageJSON(jsonObject); ok {
			return message, nil
		}
	}
	if message := extractConventionalCommit(cleaned); message != "" {
		return message, nil
	}

	return "", fmt.Errorf("could not extract commit_message from model output")
}

func firstResponsesText(resp openAICompatibleResponsesResponse) string {
	if strings.TrimSpace(resp.OutputText) != "" {
		return resp.OutputText
	}
	for _, output := range resp.Output {
		for _, content := range output.Content {
			if strings.TrimSpace(content.Text) != "" {
				return content.Text
			}
		}
	}
	return ""
}

func parseCommitMessageJSON(content string) (string, bool) {
	var parsed structuredCommitMessage
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return "", false
	}
	message := cleanMessage(parsed.CommitMessage)
	if message == "" {
		return "", false
	}
	return message, true
}

func stripModelChatter(content string) string {
	content = thinkingBlock.ReplaceAllString(content, "")
	content = strings.TrimSpace(content)
	content = strings.Trim(content, "`")

	if strings.HasPrefix(strings.ToLower(content), "json") {
		content = strings.TrimSpace(content[len("json"):])
	}

	prefixes := []string{
		"here is your commit message:",
		"here is the commit message:",
		"commit message:",
		"sure, here is the commit message:",
		"sure:",
	}
	lower := strings.ToLower(content)
	for _, prefix := range prefixes {
		if strings.HasPrefix(lower, prefix) {
			return strings.TrimSpace(content[len(prefix):])
		}
	}

	return content
}

func extractJSONObject(content string) string {
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 || end <= start {
		return ""
	}
	return content[start : end+1]
}

func extractConventionalCommit(content string) string {
	match := conventionalCommitLine.FindString(content)
	if match == "" {
		return ""
	}
	return cleanMessage(match)
}
