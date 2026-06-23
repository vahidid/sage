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
sage --provider openrouter

# choose a custom model for OpenRouter
sage --provider openrouter --model qwen/qwen3-coder:free

# show available providers
sage --list-models

# print version
sage --version
```

## Config

Create `~/.config/sage/config.json`:

```json
{
  "provider": "free",
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

Official release binaries include built-in FreeLLMApi access, so users can pick
the `free` provider without configuring an API key or model:

```bash
sage --provider free
sage --list-models
```

The free model is auto-selected as `auto` and served through FreeLLMApi.
Use the `openrouter` provider when you want to provide your own OpenRouter API
key and choose a custom model.

### Environment variables

| Variable | Description |
|----------|-------------|
| `SAGE_PROVIDER` | Override provider (`free`, `claude`, `openai`, `ollama`, `openrouter`) |
| `SAGE_DEBUG` | Show provider HTTP status and raw error details |
| `SAGE_FREE_LLM_API_KEY` | Local/dev key for built-in FreeLLMApi access |
| `SAGE_FREE_LLM_API_BASE_URL` | Local/dev FreeLLMApi base URL, for example `http://65.109.176.81:3001` |
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

Release builds can inject the bundled free key and base URL with Make arguments:

```bash
make release \
  SAGE_FREE_LLM_API_KEY=$SAGE_FREE_LLM_API_KEY \
  SAGE_FREE_LLM_API_BASE_URL=$SAGE_FREE_LLM_API_BASE_URL
```

The GitHub Actions release workflow reads the same values from repository
secrets named `SAGE_FREE_LLM_API_KEY` and `SAGE_FREE_LLM_API_BASE_URL`.

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
