# Sage — Git AI Commit

Generate concise git commit messages using AI (Claude or Ollama).
Zero external dependencies — single static binary.

## Install

```bash
git clone https://github.com/vahidid/sage.git
cd sage && make install
```

## Usage

```bash
# stage your changes first
git add -A

# generate + commit
sage

# see the message only, don't commit
sage --dry-run

# override provider for one run
sage --provider ollama

# print version
sage --version
```

## Config

Create `~/.config/sage/config.json`:

```json
{
  "provider": "claude",
  "claude": {
    "api_key": "sk-ant-...",
    "model": "claude-haiku-4-5-20251001"
  },
  "ollama": {
    "host": "http://localhost:11434",
    "model": "qwen2.5-coder:1.5b"
  }
}
```

### Environment variables

| Variable              | Description                        |
|-----------------------|------------------------------------|
| `ANTHROPIC_API_KEY`   | Claude API key (standard)          |
| `SAGE_PROVIDER`        | Override provider (claude, ollama) |
| `SAGE_CLAUDE_MODEL`    | Override Claude model              |
| `SAGE_OLLAMA_HOST`     | Override Ollama host URL           |
| `SAGE_OLLAMA_MODEL`    | Override Ollama model name         |

## Offline mode (Ollama)

```bash
# install Ollama: https://ollama.com
ollama pull qwen2.5-coder:1.5b
ollama serve

sage --provider ollama
```

## Build

```bash
make build          # current platform
make release        # all platforms → ./dist/
make install        # install to $GOPATH/bin
```
