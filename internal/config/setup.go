package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RunSetup interactively creates a config file the first time sage is used.
func RunSetup() (*Config, error) {
	fmt.Println("👋 Welcome to sage! Let's set up your configuration.")
	fmt.Println()

	cfg := defaultConfig
	r := bufio.NewReader(os.Stdin)

	// ── provider ─────────────────────────────────────────────────────────────
	fmt.Println("Choose a provider:")
	fmt.Println("  [1] Free        — built-in free models, no API key needed")
	fmt.Println("  [2] Claude      — Anthropic API key")
	fmt.Println("  [3] OpenAI      — OpenAI API key")
	fmt.Println("  [4] Ollama      — local model, no API key needed")
	fmt.Println("  [5] OpenRouter  — custom OpenRouter key/model")
	fmt.Print("\nProvider [1]: ")

	choice := strings.TrimSpace(readLine(r))
	if choice == "" {
		choice = "1"
	}

	switch choice {
	case "1", "free":
		cfg.Provider = "free"
		setupFree()
	case "2", "claude":
		cfg.Provider = "claude"
		if err := setupClaude(r, &cfg); err != nil {
			return nil, err
		}
	case "3", "openai":
		cfg.Provider = "openai"
		if err := setupOpenAI(r, &cfg); err != nil {
			return nil, err
		}
	case "4", "ollama":
		cfg.Provider = "ollama"
		setupOllama(r, &cfg)
	case "5", "openrouter":
		cfg.Provider = "openrouter"
		if err := setupOpenRouter(r, &cfg); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid choice %q — run sage again to retry", choice)
	}

	if err := saveConfig(&cfg); err != nil {
		return nil, err
	}

	fmt.Printf("\n✅ Config saved to %s\n\n", FilePath())
	return &cfg, nil
}

// Exists reports whether the config file is already on disk.
func Exists() bool {
	_, err := os.Stat(FilePath())
	return err == nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func setupFree() {
	fmt.Printf("\nFree model: %s (auto-selected)\n", DefaultFreeModel())
}

func setupClaude(r *bufio.Reader, cfg *Config) error {
	fmt.Print("\nAnthropic API key (sk-ant-...): ")
	key := strings.TrimSpace(readLine(r))
	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	cfg.Claude.APIKey = key

	fmt.Printf("Model [%s]: ", defaultConfig.Claude.Model)
	if m := strings.TrimSpace(readLine(r)); m != "" {
		cfg.Claude.Model = m
	}
	return nil
}

func setupOpenAI(r *bufio.Reader, cfg *Config) error {
	fmt.Print("\nOpenAI API key (sk-...): ")
	key := strings.TrimSpace(readLine(r))
	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	cfg.OpenAI.APIKey = key

	fmt.Printf("Model [%s]: ", defaultConfig.OpenAI.Model)
	if m := strings.TrimSpace(readLine(r)); m != "" {
		cfg.OpenAI.Model = m
	}
	return nil
}

func setupOpenRouter(r *bufio.Reader, cfg *Config) error {
	fmt.Print("\nOpenRouter API key (sk-or-...): ")
	key := strings.TrimSpace(readLine(r))
	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	cfg.OpenRouter.APIKey = key

	fmt.Printf("Model [%s]: ", defaultConfig.OpenRouter.Model)
	if m := strings.TrimSpace(readLine(r)); m != "" {
		cfg.OpenRouter.Model = m
	}
	return nil
}

func setupOllama(r *bufio.Reader, cfg *Config) {
	fmt.Printf("Ollama host [%s]: ", defaultConfig.Ollama.Host)
	if h := strings.TrimSpace(readLine(r)); h != "" {
		cfg.Ollama.Host = h
	}

	fmt.Printf("Model [%s]: ", defaultConfig.Ollama.Model)
	if m := strings.TrimSpace(readLine(r)); m != "" {
		cfg.Ollama.Model = m
	}
}

func saveConfig(cfg *Config) error {
	path := FilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("cannot create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func readLine(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimRight(line, "\r\n")
}
