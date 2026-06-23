package ai

import (
	"fmt"
	"regexp"
	"strings"
)

const maxDiffLines = 500
const commitSystemPrompt = "Return only the final commit message: one line, no explanation, no preamble, no markdown, no quotes."

// buildPrompt constructs the prompt sent to any AI provider.
// Keeping it here means all providers behave identically.
func buildPrompt(diff string) string {
	body := compactDiffForPrompt(diff)
	scope := inferScopeFromDiff(diff)
	if scope == "" {
		scope = "omit if unclear"
	}

	return fmt.Sprintf(`Write a Conventional Commit message for this git diff.
Format: <type>(<scope>): <subject>
Types: feat, fix, refactor, chore, docs, style, test, perf
Rules: <=72 chars, imperative subject, lowercase after colon, no period.
Suggested scope: %s

Examples:
diff: internal/ai/freellmapi.go
feat(ai): add freellmapi provider

diff: internal/ai/structured_output.go
fix(ai): parse structured provider responses

diff: README.md
docs(readme): document free provider setup

diff: .github/workflows/release.yml
chore(ci): inject release credentials

Diff:
%s`, scope, body)
}

// plainPrompt is the full prompt for providers that have no system role
// (Ollama): the output instruction is prepended to the shared task prompt so
// the model still knows to return only the commit message.
func plainPrompt(diff string) string {
	return commitSystemPrompt + "\n\n" + buildPrompt(diff)
}

type diffFile struct {
	path  string
	lines []string
}

func compactDiffForPrompt(diff string) string {
	files := parseDiffFiles(diff)
	if len(files) == 0 {
		return truncateDiffLines(diff)
	}

	var b strings.Builder
	b.WriteString("Files changed:\n")
	for _, file := range files {
		b.WriteString("- ")
		b.WriteString(file.path)
		b.WriteByte('\n')
	}
	b.WriteString("\nRepresentative diff:\n")

	totalLines := 0
	for _, file := range files {
		totalLines += len(file.lines)
	}

	maxPerFile := maxDiffLines / len(files)
	if maxPerFile < 24 {
		maxPerFile = 24
	}
	if maxPerFile > 120 {
		maxPerFile = 120
	}

	compacted := totalLines > maxDiffLines
	for i, file := range files {
		if i > 0 {
			b.WriteByte('\n')
		}
		limit := len(file.lines)
		if limit > maxPerFile {
			limit = maxPerFile
			compacted = true
		}
		b.WriteString(strings.Join(file.lines[:limit], "\n"))
		b.WriteByte('\n')
		if limit < len(file.lines) {
			b.WriteString(fmt.Sprintf("[file diff compacted - showing %d of %d lines]\n", limit, len(file.lines)))
		}
	}
	if compacted {
		b.WriteString(fmt.Sprintf("\n[diff compacted - showing representative hunks from %d files]", len(files)))
	}
	return strings.TrimSpace(b.String())
}

func truncateDiffLines(diff string) string {
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
	return body
}

func parseDiffFiles(diff string) []diffFile {
	var files []diffFile
	var current *diffFile
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			path := pathFromDiffHeader(line)
			files = append(files, diffFile{path: path})
			current = &files[len(files)-1]
		}
		if current != nil {
			current.lines = append(current.lines, line)
		}
	}
	return files
}

func pathFromDiffHeader(line string) string {
	parts := strings.Fields(line)
	if len(parts) >= 4 {
		path := strings.TrimPrefix(parts[3], "b/")
		if path != "/dev/null" && path != "" {
			return path
		}
		path = strings.TrimPrefix(parts[2], "a/")
		if path != "/dev/null" && path != "" {
			return path
		}
	}
	return strings.TrimSpace(strings.TrimPrefix(line, "diff --git"))
}

func inferScopeFromDiff(diff string) string {
	files := parseDiffFiles(diff)
	if len(files) == 0 {
		return ""
	}
	counts := map[string]int{}
	order := []string{}
	for _, file := range files {
		scope := scopeForPath(file.path)
		if scope == "" {
			continue
		}
		if counts[scope] == 0 {
			order = append(order, scope)
		}
		counts[scope]++
	}
	best := ""
	bestCount := 0
	for _, scope := range order {
		if counts[scope] > bestCount {
			best = scope
			bestCount = counts[scope]
		}
	}
	return best
}

func scopeForPath(path string) string {
	switch {
	case strings.HasPrefix(path, "internal/ai/"):
		return "ai"
	case strings.HasPrefix(path, "internal/config/"):
		return "config"
	case strings.HasPrefix(path, "internal/git/"):
		return "git"
	case strings.HasPrefix(path, ".github/workflows/"):
		return "ci"
	case path == "README.md":
		return "readme"
	case path == "Makefile":
		return "build"
	}
	return ""
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
	msg = stripCommitLabel(msg)
	msg = strings.Trim(msg, "\"'`")
	return strings.TrimSpace(msg)
}

// stripCommitLabel removes a leading "commit:" label that weaker models
// sometimes echo from the prompt examples.
func stripCommitLabel(msg string) string {
	const label = "commit:"
	if len(msg) >= len(label) && strings.EqualFold(msg[:len(label)], label) {
		return strings.TrimSpace(msg[len(label):])
	}
	return msg
}
