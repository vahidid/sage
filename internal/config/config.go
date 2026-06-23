package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all runtime configuration.
// Priority (highest → lowest):
//  1. Environment variables  (ANTHROPIC_API_KEY, SAGE_PROVIDER, …)
//  2. Config file            (~/.config/sage/config.json)
//  3. Built-in defaults
type Config struct {
	Provider   string           `json:"provider"`
	Claude     ClaudeConfig     `json:"claude"`
	OpenAI     OpenAIConfig     `json:"openai"`
	Ollama     OllamaConfig     `json:"ollama"`
	OpenRouter OpenRouterConfig `json:"openrouter"`
}

type ClaudeConfig struct {
	APIKey string `json:"api_key"`
	Model  string `json:"model"`
}

type OpenAIConfig struct {
	APIKey string `json:"api_key"`
	Model  string `json:"model"`
}

type OllamaConfig struct {
	Host  string `json:"host"`
	Model string `json:"model"`
}

type OpenRouterConfig struct {
	APIKey string `json:"api_key"`
	Model  string `json:"model"`
}

// defaults
var defaultConfig = Config{
	Provider: "free",
	Claude: ClaudeConfig{
		Model: "claude-haiku-4-5-20251001",
	},
	OpenAI: OpenAIConfig{
		Model: "gpt-4o-mini",
	},
	Ollama: OllamaConfig{
		Host:  "http://localhost:11434",
		Model: "qwen2.5-coder:1.5b",
	},
	OpenRouter: OpenRouterConfig{
		Model: "google/gemma-3-12b-it:free",
	},
}

// Load reads config from file and env vars, merging over defaults.
func Load() (*Config, error) {
	cfg := defaultConfig // start from defaults

	// read config file (missing file is OK)
	if data, err := os.ReadFile(FilePath()); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("invalid config file: %w", err)
		}
	}

	// env var overrides
	if v := os.Getenv("SAGE_PROVIDER"); v != "" {
		cfg.Provider = v
	}
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		cfg.Claude.APIKey = v
	}
	if v := os.Getenv("SAGE_CLAUDE_API_KEY"); v != "" {
		cfg.Claude.APIKey = v
	}
	if v := os.Getenv("SAGE_CLAUDE_MODEL"); v != "" {
		cfg.Claude.Model = v
	}
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		cfg.OpenAI.APIKey = v
	}
	if v := os.Getenv("SAGE_OPENAI_API_KEY"); v != "" {
		cfg.OpenAI.APIKey = v
	}
	if v := os.Getenv("SAGE_OPENAI_MODEL"); v != "" {
		cfg.OpenAI.Model = v
	}
	if v := os.Getenv("SAGE_OLLAMA_HOST"); v != "" {
		cfg.Ollama.Host = v
	}
	if v := os.Getenv("SAGE_OLLAMA_MODEL"); v != "" {
		cfg.Ollama.Model = v
	}
	if v := os.Getenv("OPENROUTER_API_KEY"); v != "" {
		cfg.OpenRouter.APIKey = v
	}
	if v := os.Getenv("SAGE_OPENROUTER_API_KEY"); v != "" {
		cfg.OpenRouter.APIKey = v
	}
	if v := os.Getenv("SAGE_OPENROUTER_MODEL"); v != "" {
		cfg.OpenRouter.Model = v
	}

	return &cfg, nil
}

// FilePath returns the expected config file location.
func FilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sage", "config.json")
}
