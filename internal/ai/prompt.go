package ai

import (
	"fmt"
	"regexp"
	"strings"
)

const maxDiffLines = 500
const commitSystemPrompt = "Return only one final commit message. No analysis, no reasoning, no markdown, no quotes, one line."

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
		body += "\n\n[diff truncated - showing first 500 lines]"
	}

	return fmt.Sprintf(`Write a Conventional Commit message for this git diff.
Format: <type>(<scope>): <subject>
Types: feat, fix, refactor, chore, docs, style, test, perf
Rules: <=72 chars, imperative subject, lowercase after colon, no period.

Examples:
diff: +func validateToken(token string) error
commit: feat(auth): validate tokens before use

diff: -return nil
      +return fmt.Errorf("missing config")
commit: fix(config): report missing config errors

diff: README.md
commit: docs(readme): update usage examples

Diff:
%s`, body)
}

var thinkingBlock = regexp.MustCompile(`(?s)<think(?:ing)?>.*?</think(?:ing)?>`)

// cleanMessage strips thinking tags, trims whitespace/quotes, and takes only
// the first line in case the model returns more than one line despite the prompt.
func cleanMessage(msg string) string {
	msg = thinkingBlock.ReplaceAllString(msg, "")
	msg = strings.TrimSpace(msg)
	msg = strings.Trim(msg, "\"'`")

	if idx := strings.Index(msg, "\n"); idx != -1 {
		msg = msg[:idx]
	}

	msg = strings.Trim(msg, "\"'`")
	return strings.TrimSpace(msg)
}
