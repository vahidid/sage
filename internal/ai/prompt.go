package ai

import (
	"fmt"
	"regexp"
	"strings"
)

const maxDiffLines = 500

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
		body += "\n\n[diff truncated — showing first 500 lines]"
	}

	return fmt.Sprintf(`Analyze the git diff below and output a single commit message.

## COMMIT DESIGN RULES:
- Format: <type>(<scope>): <subject>
- Types: feat, fix, refactor, chore, docs, style, test, perf
- Scope: the module, file, or domain affected (e.g. auth, billing, api)
- Subject: imperative mood — "Add X" not "Added X", "Fix Y" not "Fixed Y"
- Maximum 72 characters total
- Be specific about WHAT changed, not HOW or WHY
- No period at the end
- Lowercase after the colon

## EXAMPLES:
  feat(auth): add OAuth2 login with Google
  fix(billing): prevent double charge on retry
  refactor(api): extract pagination logic to helper
  chore(deps): upgrade Laravel to 11.x

## OUTPUT RULE:
Return ONLY the commit message — one line, no quotes, no backticks, no explanation, no prefix like "Commit:" or "Message:". Any extra word is a failure.

Git diff:
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

	return strings.TrimSpace(msg)
}
