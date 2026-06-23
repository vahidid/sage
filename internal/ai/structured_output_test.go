package ai

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseStructuredCommitMessageReturnsOnlyCommitMessage(t *testing.T) {
	got, err := parseStructuredCommitMessage(`{"commit_message":"fix(api): return status"}`)
	if err != nil {
		t.Fatalf("parseStructuredCommitMessage() error = %v", err)
	}
	if got != "fix(api): return status" {
		t.Fatalf("parseStructuredCommitMessage() = %q, want %q", got, "fix(api): return status")
	}
}

func TestParseStructuredCommitMessageCleansCommitMessageValue(t *testing.T) {
	got, err := parseStructuredCommitMessage(`{"commit_message":"\"fix(api): return status\"\nextra"}`)
	if err != nil {
		t.Fatalf("parseStructuredCommitMessage() error = %v", err)
	}
	if got != "fix(api): return status" {
		t.Fatalf("parseStructuredCommitMessage() = %q, want %q", got, "fix(api): return status")
	}
}

func TestParseStructuredCommitMessageExtractsFromCodeFence(t *testing.T) {
	got, err := parseStructuredCommitMessage("```json\n{\"commit_message\":\"fix(api): return status\"}\n```")
	if err != nil {
		t.Fatalf("parseStructuredCommitMessage() error = %v", err)
	}
	if got != "fix(api): return status" {
		t.Fatalf("parseStructuredCommitMessage() = %q, want %q", got, "fix(api): return status")
	}
}

func TestParseStructuredCommitMessageExtractsConventionalCommitFallback(t *testing.T) {
	got, err := parseStructuredCommitMessage("Here is the commit message:\nfix(api): return status\n\nThis updates API handling.")
	if err != nil {
		t.Fatalf("parseStructuredCommitMessage() error = %v", err)
	}
	if got != "fix(api): return status" {
		t.Fatalf("parseStructuredCommitMessage() = %q, want %q", got, "fix(api): return status")
	}
}

func TestParseStructuredCommitMessageRejectsUnusableOutput(t *testing.T) {
	if _, err := parseStructuredCommitMessage("I changed the API and config files."); err == nil {
		t.Fatal("parseStructuredCommitMessage() error = nil, want non-nil")
	}
}

func TestParseStructuredCommitMessageRejectsInvalidCommitMessage(t *testing.T) {
	if _, err := parseStructuredCommitMessage(`{"commit_message":"updated the API and config files"}`); err == nil {
		t.Fatal("parseStructuredCommitMessage() error = nil, want non-nil")
	}
}

func TestParseProviderCommitMessageAcceptsStructuredJSON(t *testing.T) {
	got, err := parseProviderCommitMessage(`{"commit_message":"fix(api): return status"}`)
	if err != nil {
		t.Fatalf("parseProviderCommitMessage() error = %v", err)
	}
	if got != "fix(api): return status" {
		t.Fatalf("parseProviderCommitMessage() = %q", got)
	}
}

func TestValidateCommitMessage(t *testing.T) {
	valid := []string{
		"fix(api): return status",
		"docs(readme): document free provider",
		"chore(ci): inject release credentials",
	}
	for _, msg := range valid {
		if err := validateCommitMessage(msg); err != nil {
			t.Fatalf("validateCommitMessage(%q) error = %v", msg, err)
		}
	}

	invalid := []string{
		"updated the API",
		"fix(api): Return status",
		"fix(api): return status.",
		"fix(api): return status\nextra",
		"build(api): return status",
	}
	for _, msg := range invalid {
		if err := validateCommitMessage(msg); err == nil {
			t.Fatalf("validateCommitMessage(%q) error = nil, want non-nil", msg)
		}
	}
}

func TestCommitMessageTextUsesStrictJSONSchema(t *testing.T) {
	raw, err := json.Marshal(commitMessageText())
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	body := string(raw)

	checks := []string{
		`"type":"json_schema"`,
		`"name":"commit_message_response"`,
		`"strict":true`,
		`"commit_message"`,
		`"additionalProperties":false`,
	}
	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Fatalf("response format missing %s in %s", check, body)
		}
	}
}

func TestFirstResponsesTextUsesOutputText(t *testing.T) {
	got := firstResponsesText(openAICompatibleResponsesResponse{
		OutputText: `{"commit_message":"fix(api): return status"}`,
	})
	if got != `{"commit_message":"fix(api): return status"}` {
		t.Fatalf("firstResponsesText() = %q", got)
	}
}

func TestFirstResponsesTextFallsBackToOutputContent(t *testing.T) {
	resp := openAICompatibleResponsesResponse{
		Output: []openAICompatibleOutputItem{
			{
				Type: "message",
				Content: []openAICompatibleOutputContent{
					{Type: "output_text", Text: `{"commit_message":"fix(api): return status"}`},
				},
			},
		},
	}

	got := firstResponsesText(resp)
	if got != `{"commit_message":"fix(api): return status"}` {
		t.Fatalf("firstResponsesText() = %q", got)
	}
}

func TestFirstResponsesTextSkipsReasoningItem(t *testing.T) {
	resp := openAICompatibleResponsesResponse{
		Output: []openAICompatibleOutputItem{
			{Type: "reasoning"},
			{
				Type: "message",
				Content: []openAICompatibleOutputContent{
					{Type: "output_text", Text: `{"commit_message":"fix(api): return status"}`},
				},
			},
		},
	}

	if got := firstResponsesText(resp); got != `{"commit_message":"fix(api): return status"}` {
		t.Fatalf("firstResponsesText() = %q, want the message item text", got)
	}
}

func TestExtractResponsesTextReportsTruncation(t *testing.T) {
	resp := openAICompatibleResponsesResponse{
		Status:     "incomplete",
		Incomplete: &openAICompatibleIncomplete{Reason: "max_output_tokens"},
	}

	_, err := extractResponsesText("OpenAI", resp, 200, []byte(`{"status":"incomplete"}`))
	if err == nil {
		t.Fatal("extractResponsesText() error = nil, want truncation error")
	}
	if !strings.Contains(err.Error(), "cut off") {
		t.Fatalf("extractResponsesText() error = %q, want truncation message", err)
	}
}

func TestNewResponsesRequestEnablesReasoningAndOmitsTemperature(t *testing.T) {
	raw, err := json.Marshal(newOpenAICompatibleResponsesRequest("m", nil))
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	body := string(raw)

	if !strings.Contains(body, `"reasoning":{"effort":"low"}`) {
		t.Fatalf("request missing reasoning config: %s", body)
	}
	if !strings.Contains(body, `"max_output_tokens":2048`) {
		t.Fatalf("request missing token budget: %s", body)
	}
	if strings.Contains(body, `"temperature"`) {
		t.Fatalf("request should omit temperature for reasoning models: %s", body)
	}
}
