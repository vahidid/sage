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

func TestCleanMessageStripsCommitLabel(t *testing.T) {
	cases := []struct {
		name string
		msg  string
		want string
	}{
		{name: "lowercase", msg: "commit: fix(api): return status", want: "fix(api): return status"},
		{name: "capitalized", msg: "Commit: fix(api): return status", want: "fix(api): return status"},
		{name: "quoted after label", msg: `commit: "fix(api): return status"`, want: "fix(api): return status"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := cleanMessage(tc.msg); got != tc.want {
				t.Fatalf("cleanMessage() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCommitSystemPromptForbidsExplanations(t *testing.T) {
	if !strings.Contains(strings.ToLower(commitSystemPrompt), "no explanation") {
		t.Fatalf("commitSystemPrompt should forbid explanations: %q", commitSystemPrompt)
	}
}

func TestPlainPromptPrependsOutputInstruction(t *testing.T) {
	diff := "diff --git a/a b/a\n+change"
	got := plainPrompt(diff)

	if !strings.HasPrefix(got, commitSystemPrompt) {
		t.Fatalf("plainPrompt() should start with the output instruction:\n%s", got)
	}
	if !strings.Contains(got, diff) {
		t.Fatalf("plainPrompt() missing diff body:\n%s", got)
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

func TestBuildPromptIncludesFileSummaryAndSuggestedScope(t *testing.T) {
	diff := strings.Join([]string{
		"diff --git a/internal/ai/openai.go b/internal/ai/openai.go",
		"@@ -1,3 +1,4 @@",
		"+func generateStructuredCommit() {}",
		"diff --git a/internal/ai/prompt_test.go b/internal/ai/prompt_test.go",
		"@@ -1,3 +1,4 @@",
		"+func TestPromptQuality(t *testing.T) {}",
	}, "\n")

	got := buildPrompt(diff)

	checks := []string{
		"Suggested scope: ai",
		"Files changed:",
		"- internal/ai/openai.go",
		"- internal/ai/prompt_test.go",
		"Representative diff:",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("buildPrompt() missing %q:\n%s", check, got)
		}
	}
}

func TestBuildPromptSamplesAcrossFilesWhenLargeDiffIsCompacted(t *testing.T) {
	var firstFile []string
	firstFile = append(firstFile, "diff --git a/internal/ai/large.go b/internal/ai/large.go", "@@ -1,3 +1,700 @@")
	for i := 0; i < maxDiffLines+200; i++ {
		firstFile = append(firstFile, fmt.Sprintf("+large change %03d", i))
	}
	diff := strings.Join(append(firstFile,
		"diff --git a/README.md b/README.md",
		"@@ -1,3 +1,4 @@",
		"+document the free provider",
	), "\n")

	got := buildPrompt(diff)

	if !strings.Contains(got, "- README.md") {
		t.Fatalf("buildPrompt() should include later changed files after compaction:\n%s", got)
	}
	if !strings.Contains(got, "+document the free provider") {
		t.Fatalf("buildPrompt() should sample representative hunks from later files:\n%s", got)
	}
	if !strings.Contains(got, "[diff compacted") {
		t.Fatalf("buildPrompt() missing compaction marker:\n%s", got)
	}
}

func TestInferScopeFromDiff(t *testing.T) {
	tests := []struct {
		name string
		diff string
		want string
	}{
		{name: "ai", diff: "diff --git a/internal/ai/openai.go b/internal/ai/openai.go", want: "ai"},
		{name: "config", diff: "diff --git a/internal/config/config.go b/internal/config/config.go", want: "config"},
		{name: "readme", diff: "diff --git a/README.md b/README.md", want: "readme"},
		{name: "ci", diff: "diff --git a/.github/workflows/release.yml b/.github/workflows/release.yml", want: "ci"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inferScopeFromDiff(tt.diff); got != tt.want {
				t.Fatalf("inferScopeFromDiff() = %q, want %q", got, tt.want)
			}
		})
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
