package ai

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const structuredCommitSystemPrompt = "Return a JSON object that matches the provided schema. The commit_message value must be a single Conventional Commit message with no explanation, preamble, or markdown."

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
	Reasoning       *reasoningConfig          `json:"reasoning,omitempty"`
	// Temperature is a pointer so it can be omitted: reasoning models reject it.
	Temperature *float64 `json:"temperature,omitempty"`
}

// reasoningConfig enables provider-side reasoning on the Responses API.
// Effort is fixed to "low" for now; it will become configurable later.
type reasoningConfig struct {
	Effort string `json:"effort"`
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
	OutputText string                       `json:"output_text,omitempty"`
	Status     string                       `json:"status,omitempty"`
	Incomplete *openAICompatibleIncomplete  `json:"incomplete_details,omitempty"`
	Output     []openAICompatibleOutputItem `json:"output,omitempty"`
	Error      *providerAPIError            `json:"error,omitempty"`
}

type openAICompatibleIncomplete struct {
	Reason string `json:"reason"`
}

// openAICompatibleOutputItem is one item in the Responses API output array.
// With reasoning enabled the array holds a "reasoning" item and a "message"
// item; only the latter carries the answer.
type openAICompatibleOutputItem struct {
	Type    string                          `json:"type"`
	Content []openAICompatibleOutputContent `json:"content,omitempty"`
}

type openAICompatibleOutputContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
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

const (
	// maxCommitTokens must cover reasoning tokens plus the JSON answer, since
	// the Responses API counts reasoning against this budget.
	maxCommitTokens = 2048
	// reasoningEffort is fixed for now; it will become configurable later.
	reasoningEffort = "low"
)

func newOpenAICompatibleResponsesRequest(model string, input []openAICompatibleMessage) openAICompatibleResponsesRequest {
	return openAICompatibleResponsesRequest{
		Model:           model,
		Input:           input,
		Text:            commitMessageText(),
		MaxOutputTokens: maxCommitTokens,
		Reasoning:       &reasoningConfig{Effort: reasoningEffort},
		// Temperature is intentionally omitted: reasoning models reject it.
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

func parseStructuredCommitMessage(content string) (string, error) {
	// Providers with strict json_schema return a clean JSON object.
	if message, ok := parseCommitMessageJSON(content); ok {
		return message, nil
	}

	// Free models sometimes wrap the JSON in prose or markdown fences;
	// pull out the JSON object and try once more.
	cleaned := stripModelChatter(content)
	if jsonObject := extractJSONObject(cleaned); jsonObject != "" {
		if message, ok := parseCommitMessageJSON(jsonObject); ok {
			return message, nil
		}
	}

	// Last resort: grab a Conventional Commit line directly from the text.
	if message := extractConventionalCommit(cleaned); message != "" {
		return message, nil
	}

	return "", fmt.Errorf("could not extract commit_message from model output")
}

// extractResponsesText returns the assistant message text, or a descriptive
// error when the response is empty or was truncated before completing (which
// usually means reasoning exhausted maxCommitTokens).
func extractResponsesText(provider string, resp openAICompatibleResponsesResponse, status int, raw []byte) (string, error) {
	if content := firstResponsesText(resp); content != "" {
		return content, nil
	}
	if resp.Status == "incomplete" && resp.Incomplete != nil && resp.Incomplete.Reason == "max_output_tokens" {
		return "", formatProviderTruncated(provider, status, raw)
	}
	return "", formatProviderEmptyResponse(provider, status, raw)
}

func firstResponsesText(resp openAICompatibleResponsesResponse) string {
	if strings.TrimSpace(resp.OutputText) != "" {
		return resp.OutputText
	}
	for _, output := range resp.Output {
		// With reasoning enabled the answer lives in the "message" item;
		// skip "reasoning" (and any other) item types.
		if output.Type != "" && output.Type != "message" {
			continue
		}
		for _, content := range output.Content {
			if content.Type != "" && content.Type != "output_text" {
				continue
			}
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

// stripModelChatter removes thinking blocks and markdown code fences so the
// JSON object or Conventional Commit line underneath can be extracted.
func stripModelChatter(content string) string {
	content = thinkingBlock.ReplaceAllString(content, "")
	content = strings.TrimSpace(content)
	return strings.Trim(content, "`")
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
