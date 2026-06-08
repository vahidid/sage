package ai

// Provider is the interface every AI backend must implement.
// Adding a new provider (OpenAI, Gemini, local LLM, ...) means
// implementing these two methods — nothing else changes.
type Provider interface {
	// GenerateCommitMessage receives a git diff and returns a commit message.
	GenerateCommitMessage(diff string) (string, error)

	// Name returns the provider label used in logs ("claude", "ollama", …).
	Name() string
}
