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

// Commit runs `git commit -m <message>`.
func Commit(message string) error {
	out, err := exec.Command("git", "commit", "-m", message).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed:\n%s", string(out))
	}
	return nil
}
