package ai

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatProviderAPIErrorHidesDebugDetailsByDefault(t *testing.T) {
	t.Setenv("SAGE_DEBUG", "")

	err := formatProviderAPIError("OpenRouter", 502, providerAPIError{
		Message:  "Provider returned error",
		Metadata: json.RawMessage(`{"provider_name":"Example","raw":{"error":"rate limited"}}`),
	}, []byte(`{"error":{"message":"Provider returned error"}}`))

	if got := err.Error(); got != "OpenRouter API error: Provider returned error" {
		t.Fatalf("error = %q", got)
	}
}

func TestFormatProviderAPIErrorShowsDebugDetails(t *testing.T) {
	t.Setenv("SAGE_DEBUG", "1")

	err := formatProviderAPIError("OpenRouter", 502, providerAPIError{
		Message:  "Provider returned error",
		Code:     json.RawMessage(`502`),
		Metadata: json.RawMessage(`{"provider_name":"Example","raw":{"error":"rate limited"}}`),
	}, []byte(`{"error":{"message":"Provider returned error"}}`))

	got := err.Error()
	for _, want := range []string{
		"OpenRouter API error: Provider returned error",
		"status=502",
		"code=502",
		`metadata={"provider_name":"Example","raw":{"error":"rate limited"}}`,
		`raw_response={"error":{"message":"Provider returned error"}}`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("error missing %q:\n%s", want, got)
		}
	}
}
