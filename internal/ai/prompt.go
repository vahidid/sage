package ai

import (
	"fmt"
	"strings"
)

const maxDiffLines = 300

// buildPrompt constructs the prompt sent to any AI provider.
// Keeping it here means all providers behave identically.
func buildPrompt(diff string) string {
	lines := strings.Split(diff, "\n")
	truncated := false

	if len(lines) > maxDiffLines {
		lines = lines[:maxDiffLines]
		truncated = true
	}

	body := strings.Join(lines, "\n")
	if truncated {
		body += "\n\n[diff truncated — showing first 300 lines]"
	}

	return fmt.Sprintf(`Analyze this git diff and write a commit message.

Rules:
- Maximum 72 characters
- Imperative mood: "Fix bug" not "Fixed bug"
- Be specific about WHAT changed, not HOW
- Output ONLY the commit message — no explanation, no quotes, no formatting

Git diff:
%s`, body)
}

// cleanMessage trims whitespace, stray quotes, and takes only the first line
// in case the model returns more than one line despite the prompt.
func cleanMessage(msg string) string {
	msg = strings.TrimSpace(msg)
	msg = strings.Trim(msg, "\"'`")

	if idx := strings.Index(msg, "\n"); idx != -1 {
		msg = msg[:idx]
	}

	return strings.TrimSpace(msg)
}
