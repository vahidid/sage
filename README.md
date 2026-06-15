# Sage — Git AI Commit

Generate concise git commit messages using AI.
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
sage --provider free
sage --provider ollama

# choose a model for one run
sage --provider free --model 2
sage --provider free --model qwen/qwen3-coder:free

# show built-in free models
sage --list-models

# print version
sage --version
```

## Config

Create `~/.config/sage/config.json`:

```json
{
  "provider": "free",
  "free": {
    "model": "qwen/qwen3-coder:free"
  },
  "claude": {
    "api_key": "sk-ant-...",
    "model": "claude-haiku-4-5-20251001"
  },
  "openai": {
    "api_key": "sk-...",
    "model": "gpt-4o-mini"
  },
  "ollama": {
    "host": "http://localhost:11434",
    "model": "qwen2.5-coder:1.5b"
  },
  "openrouter": {
    "api_key": "sk-or-...",
    "model": "qwen/qwen3-coder:free"
  }
}
```

### Built-in free models

Official release binaries can include a built-in OpenRouter key, so users can
pick the `free` provider without configuring their own API key:

```bash
sage --provider free
sage --list-models
```

The bundled free presets are:

| Model | Use case |
|-------|----------|
| `qwen/qwen3-coder:free` | code-heavy diffs |
| `openai/gpt-oss-20b:free` | fast general-purpose commits |
| `google/gemma-4-26b-a4b-it:free` | balanced open model |
| `meta-llama/llama-3.3-70b-instruct:free` | strong general instruction following |

For the `free` provider, `--model` accepts either the model ID or the number
shown by `sage --list-models`.

### Environment variables

| Variable | Description |
|----------|-------------|
| `SAGE_PROVIDER` | Override provider (`free`, `claude`, `openai`, `ollama`, `openrouter`) |
| `SAGE_DEBUG` | Show provider HTTP status and raw error details |
| `SAGE_FREE_MODEL` | Override built-in free model |
| `SAGE_FREE_OPENROUTER_API_KEY` | Local/dev key for built-in free models |
| `ANTHROPIC_API_KEY` | Claude API key (standard) |
| `SAGE_CLAUDE_API_KEY` | Claude API key |
| `SAGE_CLAUDE_MODEL` | Override Claude model |
| `OPENAI_API_KEY` | OpenAI API key (standard) |
| `SAGE_OPENAI_API_KEY` | OpenAI API key |
| `SAGE_OPENAI_MODEL` | Override OpenAI model |
| `SAGE_OLLAMA_HOST` | Override Ollama host URL |
| `SAGE_OLLAMA_MODEL` | Override Ollama model name |
| `OPENROUTER_API_KEY` | OpenRouter API key (standard) |
| `SAGE_OPENROUTER_API_KEY` | OpenRouter API key |
| `SAGE_OPENROUTER_MODEL` | Override OpenRouter model |

Release builds can inject the bundled free key with:

```bash
go build -ldflags="-s -w -X main.BuiltinOpenRouterAPIKey=$SAGE_FREE_OPENROUTER_API_KEY" .
```

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
