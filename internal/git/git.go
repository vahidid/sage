package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// IsGitRepo returns true if the current directory is inside a git repository.
func IsGitRepo() bool {
	return exec.Command("git", "rev-parse", "--git-dir").Run() == nil
}

// GetStagedDiff returns the diff of all staged changes.
// Returns a descriptive error if nothing is staged.
func GetStagedDiff() (string, error) {
	out, err := exec.Command("git", "diff", "--staged").Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}

	fmt.Printf("Test For free models:")

	diff := strings.TrimSpace(string(out))
	if diff == "" {
		return "", fmt.Errorf(
			"❌ no staged changes\n" +
				"   Stage your files first:  git add <file>\n" +
				"   Or stage everything:     git add -A",
		)
	}

	return diff, nil
}

// StageAll runs `git add -A` to stage all tracked and untracked changes.
func StageAll() error {
	out, err := exec.Command("git", "add", "-A").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed:\n%s", string(out))
	}
	return nil
}

// HasUnstagedChanges reports whether there are any unstaged modifications.
func HasUnstagedChanges() bool {
	out, _ := exec.Command("git", "status", "--porcelain").Output()
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 2 {
			continue
		}
		// first char = index (staged), second char = worktree (unstaged)
		if line[1] != ' ' && line[1] != '?' {
			return true
		}
		// untracked files (both chars are '?')
		if line[0] == '?' && line[1] == '?' {
			return true
		}
	}
	return false
}

// Commit runs `git commit -m <message>`.
func Commit(message string) error {
	out, err := exec.Command("git", "commit", "-m", message).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed:\n%s", string(out))
	}
	return nil
}
