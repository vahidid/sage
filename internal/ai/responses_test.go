package ai

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateViaResponsesRetriesInvalidCommitMessageOnce(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "application/json")
		if attempts == 1 {
			fmt.Fprint(w, `{"output_text":"{\"commit_message\":\"updated the API\"}"}`)
			return
		}
		fmt.Fprint(w, `{"output_text":"{\"commit_message\":\"fix(api): return status\"}"}`)
	}))
	defer server.Close()

	got, err := generateViaResponses("TestProvider", server.URL, "test-key", "test-model", "diff --git a/internal/ai/openai.go b/internal/ai/openai.go\n+change", nil)
	if err != nil {
		t.Fatalf("generateViaResponses() error = %v", err)
	}
	if got != "fix(api): return status" {
		t.Fatalf("generateViaResponses() = %q", got)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}
