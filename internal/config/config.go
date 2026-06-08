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
	Provider string       `json:"provider"`
	Claude   ClaudeConfig `json:"claude"`
	Ollama   OllamaConfig `json:"ollama"`
}

type ClaudeConfig struct {
	APIKey string `json:"api_key"`
	Model  string `json:"model"`
}

type OllamaConfig struct {
	Host  string `json:"host"`
	Model string `json:"model"`
}

// defaults
var defaultConfig = Config{
	Provider: "claude",
	Claude: ClaudeConfig{
		Model: "claude-haiku-4-5-20251001",
	},
	Ollama: OllamaConfig{
		Host:  "http://localhost:11434",
		Model: "qwen2.5-coder:1.5b",
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
	if v := os.Getenv("SAGE_OLLAMA_HOST"); v != "" {
		cfg.Ollama.Host = v
	}
	if v := os.Getenv("SAGE_OLLAMA_MODEL"); v != "" {
		cfg.Ollama.Model = v
	}

	return &cfg, nil
}

// FilePath returns the expected config file location.
func FilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sage", "config.json")
}
