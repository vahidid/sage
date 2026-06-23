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

func TestParseStructuredCommitMessageHandlesAssistantPrefill(t *testing.T) {
	got, err := parseStructuredCommitMessage(`"commit_message":"fix(api): return status"}`)
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
	resp := openAICompatibleResponsesResponse{}
	resp.Output = append(resp.Output, struct {
		Content []struct {
			Text string `json:"text,omitempty"`
		} `json:"content,omitempty"`
	}{
		Content: []struct {
			Text string `json:"text,omitempty"`
		}{
			{Text: `{"commit_message":"fix(api): return status"}`},
		},
	})

	got := firstResponsesText(resp)
	if got != `{"commit_message":"fix(api): return status"}` {
		t.Fatalf("firstResponsesText() = %q", got)
	}
}
