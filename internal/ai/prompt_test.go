package ai

import (
	"fmt"
	"strings"
	"testing"
)

func TestCleanMessageStripsThinkingBlocks(t *testing.T) {
	got := cleanMessage("<think>consider options</think>\nfix(api): return status")
	want := "fix(api): return status"

	if got != want {
		t.Fatalf("cleanMessage() = %q, want %q", got, want)
	}
}

func TestCleanMessageStripsThinkingTagVariant(t *testing.T) {
	got := cleanMessage("<thinking>consider options</thinking>\nfix(api): return status")
	want := "fix(api): return status"

	if got != want {
		t.Fatalf("cleanMessage() = %q, want %q", got, want)
	}
}

func TestCleanMessageTrimsQuotesAndBackticks(t *testing.T) {
	cases := []struct {
		name string
		msg  string
		want string
	}{
		{name: "double quotes", msg: `"fix(api): return status"`, want: "fix(api): return status"},
		{name: "single quotes", msg: `'fix(api): return status'`, want: "fix(api): return status"},
		{name: "backticks", msg: "`fix(api): return status`", want: "fix(api): return status"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := cleanMessage(tc.msg); got != tc.want {
				t.Fatalf("cleanMessage() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCleanMessageReturnsOnlyFirstLine(t *testing.T) {
	got := cleanMessage("fix(api): return status\nextra explanation")
	want := "fix(api): return status"

	if got != want {
		t.Fatalf("cleanMessage() = %q, want %q", got, want)
	}
}

func TestBuildPromptIncludesDiffBody(t *testing.T) {
	diff := "diff --git a/main.go b/main.go\n+fmt.Println(\"hello\")"
	got := buildPrompt(diff)

	if !strings.Contains(got, diff) {
		t.Fatalf("buildPrompt() did not include diff body:\n%s", got)
	}
}

func TestBuildPromptTruncatesAfterMaxDiffLines(t *testing.T) {
	lines := make([]string, maxDiffLines+1)
	for i := range lines {
		lines[i] = fmt.Sprintf("line-%03d", i)
	}
	lines[maxDiffLines-1] = "last-included-line"
	lines[maxDiffLines] = "first-excluded-line"

	got := buildPrompt(strings.Join(lines, "\n"))

	if !strings.Contains(got, "last-included-line") {
		t.Fatalf("buildPrompt() missing final included line")
	}
	if strings.Contains(got, "first-excluded-line") {
		t.Fatalf("buildPrompt() included a line after the truncation limit")
	}
	if !strings.Contains(got, "[diff truncated - showing first 500 lines]") {
		t.Fatalf("buildPrompt() missing truncation marker")
	}
}

func TestBuildPromptIsConcise(t *testing.T) {
	got := buildPrompt("diff --git a/a b/a\n+change")

	if !strings.Contains(got, "Write a Conventional Commit message") {
		t.Fatalf("buildPrompt() missing compact instruction")
	}
	if strings.Contains(got, "COMMIT DESIGN RULES") || strings.Contains(got, "OUTPUT RULE") {
		t.Fatalf("buildPrompt() still contains verbose legacy sections")
	}
}
