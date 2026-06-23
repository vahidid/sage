package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type providerAPIError struct {
	Message  string          `json:"message"`
	Type     string          `json:"type,omitempty"`
	Code     json.RawMessage `json:"code,omitempty"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

func debugEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("SAGE_DEBUG"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func formatProviderAPIError(provider string, status int, apiErr providerAPIError, raw []byte) error {
	message := apiErr.Message
	if message == "" {
		message = "request failed"
	}

	if !debugEnabled() {
		return fmt.Errorf("%s API error: %s", provider, message)
	}

	var details []string
	if status > 0 {
		details = append(details, fmt.Sprintf("status=%d", status))
	}
	if apiErr.Type != "" {
		details = append(details, "type="+apiErr.Type)
	}
	if code := compactJSON(apiErr.Code); code != "" {
		details = append(details, "code="+code)
	}
	if metadata := compactJSON(apiErr.Metadata); metadata != "" {
		details = append(details, "metadata="+metadata)
	}
	if rawResponse := compactJSON(raw); rawResponse != "" {
		details = append(details, "raw_response="+rawResponse)
	}

	if len(details) == 0 {
		return fmt.Errorf("%s API error: %s", provider, message)
	}
	return fmt.Errorf("%s API error: %s\n   debug: %s", provider, message, strings.Join(details, "; "))
}

func formatProviderParseError(provider string, status int, raw []byte, err error) error {
	if !debugEnabled() {
		return fmt.Errorf("failed to parse %s response: %w", provider, err)
	}
	return fmt.Errorf(
		"failed to parse %s response: %w\n   debug: status=%d; raw_response=%s",
		provider,
		err,
		status,
		compactJSON(raw),
	)
}

func formatProviderTruncated(provider string, status int, raw []byte) error {
	if !debugEnabled() {
		return fmt.Errorf("%s response was cut off before completing (the model ran out of output budget) — please try again", provider)
	}
	return fmt.Errorf(
		"%s response was cut off before completing (the model ran out of output budget) — please try again\n   debug: status=%d; raw_response=%s",
		provider,
		status,
		compactJSON(raw),
	)
}

func formatProviderEmptyResponse(provider string, status int, raw []byte) error {
	if !debugEnabled() {
		return fmt.Errorf("empty response from %s", provider)
	}
	return fmt.Errorf("empty response from %s\n   debug: status=%d; raw_response=%s", provider, status, compactJSON(raw))
}

// debugLogRequest prints the outgoing request body to stderr when SAGE_DEBUG
// is enabled, so you can inspect exactly what is sent to a provider.
func debugLogRequest(provider, url string, body []byte) {
	if !debugEnabled() {
		return
	}
	fmt.Fprintf(os.Stderr, "🐛 [%s] → POST %s\n   request: %s\n", provider, url, debugPayload(body))
}

// debugLogResponse prints the raw response (and status) to stderr when
// SAGE_DEBUG is enabled, so you can inspect exactly what a provider returned.
func debugLogResponse(provider string, status int, raw []byte) {
	if !debugEnabled() {
		return
	}
	fmt.Fprintf(os.Stderr, "🐛 [%s] ← status=%d\n   response: %s\n", provider, status, debugPayload(raw))
}

// debugPayload renders a payload compactly for logging, falling back to a
// placeholder when it is empty.
func debugPayload(raw []byte) string {
	if compacted := compactJSON(raw); compacted != "" {
		return compacted
	}
	return "(empty)"
}

func compactJSON(raw []byte) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}

	var compacted bytes.Buffer
	if err := json.Compact(&compacted, raw); err != nil {
		return strings.TrimSpace(string(raw))
	}
	return compacted.String()
}
