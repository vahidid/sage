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

func formatProviderEmptyResponse(provider string, status int, raw []byte) error {
	if !debugEnabled() {
		return fmt.Errorf("empty response from %s", provider)
	}
	return fmt.Errorf("empty response from %s\n   debug: status=%d; raw_response=%s", provider, status, compactJSON(raw))
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
