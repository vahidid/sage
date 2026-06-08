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
	fmt.Println("  [1] Claude  — Anthropic API (recommended)")
	fmt.Println("  [2] Ollama  — local model, no API key needed")
	fmt.Print("\nProvider [1]: ")

	choice := strings.TrimSpace(readLine(r))
	if choice == "" {
		choice = "1"
	}

	switch choice {
	case "1", "claude":
		cfg.Provider = "claude"
		if err := setupClaude(r, &cfg); err != nil {
			return nil, err
		}
	case "2", "ollama":
		cfg.Provider = "ollama"
		setupOllama(r, &cfg)
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
