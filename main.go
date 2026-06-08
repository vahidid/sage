package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vahidid/sage/internal/ai"
	"github.com/vahidid/sage/internal/config"
	"github.com/vahidid/sage/internal/git"
)

const version = "0.1.0"

func main() {
	// ── flags ─────────────────────────────────────────────────────────────────
	dryRun   := flag.Bool("dry-run", false, "Generate message without committing")
	provider := flag.String("provider", "", "Override provider (claude, ollama)")
	ver      := flag.Bool("version", false, "Print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `sage %s — Git AI Commit

Usage:
  sage [flags]

Flags:
`, version)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  sage                        generate and commit
  sage --dry-run              generate only, no commit
  sage --provider ollama      use local Ollama model

Config file: ~/.config/sage/config.json
Docs:        github.com/vahidid/sage
`)
	}

	flag.Parse()

	if *ver {
		fmt.Println("sage", version)
		return
	}

	if err := run(*dryRun, *provider); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(dryRun bool, providerOverride string) error {
	// 1. must be inside a git repo
	if !git.IsGitRepo() {
		return fmt.Errorf("❌ not a git repository")
	}

	// 2. load config (run first-time setup if no config file exists)
	var (
		cfg *config.Config
		err error
	)
	if !config.Exists() {
		cfg, err = config.RunSetup()
		if err != nil {
			return fmt.Errorf("setup failed: %w", err)
		}
	} else {
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
	}
	if providerOverride != "" {
		cfg.Provider = providerOverride
	}

	// 3. staged diff
	fmt.Println("📋 Reading staged changes...")
	diff, err := git.GetStagedDiff()
	if err != nil {
		return err
	}

	// 4. resolve provider
	p, err := resolveProvider(cfg)
	if err != nil {
		return err
	}
	fmt.Printf("🤖 Generating with %s...\n", p.Name())

	// 5. generate
	message, err := p.GenerateCommitMessage(diff)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// 6. display
	fmt.Printf("\n  💬 %s\n\n", message)

	if dryRun {
		fmt.Println("(dry-run — no commit made)")
		return nil
	}

	// 7. confirm
	if !confirm("Commit with this message? [Y/n]: ") {
		fmt.Println("Aborted.")
		return nil
	}

	// 8. commit
	if err := git.Commit(message); err != nil {
		return err
	}
	fmt.Println("✅ Committed!")
	return nil
}

func resolveProvider(cfg *config.Config) (ai.Provider, error) {
	switch cfg.Provider {
	case "claude":
		if cfg.Claude.APIKey == "" {
			return nil, fmt.Errorf(
				"❌ Claude API key not set\n"+
					"   Set env var:  export ANTHROPIC_API_KEY=sk-ant-...\n"+
					"   Or add it to: %s", config.FilePath(),
			)
		}
		return ai.NewClaudeProvider(cfg.Claude.APIKey, cfg.Claude.Model), nil

	case "ollama":
		return ai.NewOllamaProvider(cfg.Ollama.Host, cfg.Ollama.Model), nil

	default:
		return nil, fmt.Errorf("❌ unknown provider %q — choose: claude, ollama", cfg.Provider)
	}
}

func confirm(prompt string) bool {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return input == "" || input == "y" || input == "Y" || input == "yes"
}
