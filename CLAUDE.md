# gac — Git AI Commit

## Project Context
A CLI tool written in Go that generates git commit messages using AI.
Zero external dependencies — pure stdlib.

## Architecture
- `main.go` — entry point, flag parsing, orchestration
- `internal/ai/provider.go` — Provider interface (core abstraction)
- `internal/ai/claude.go` — Anthropic API provider
- `internal/ai/ollama.go` — Local Ollama provider
- `internal/ai/prompt.go` — shared prompt builder
- `internal/git/git.go` — thin wrapper around git CLI
- `internal/config/config.go` — JSON config + env vars

## Commands
- `go build -o gac .` — build binary
- `make release` — cross-platform builds → ./dist/
- `make install` — install to $GOPATH/bin

## Adding a new provider
Implement the `Provider` interface in `internal/ai/`:
  - `Name() string`
  - `GenerateCommitMessage(diff string) (string, error)`
Then add a case to `resolveProvider()` in `main.go`.