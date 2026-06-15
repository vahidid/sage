package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vahidid/sage/internal/ai"
	"github.com/vahidid/sage/internal/config"
	"github.com/vahidid/sage/internal/git"
)

var Version = "0.1.0"
var BuiltinOpenRouterAPIKey = ""

func main() {
	// ── flags ─────────────────────────────────────────────────────────────────
	dryRun := flag.Bool("dry-run", false, "Generate message without committing")
	stageAll := flag.Bool("all", false, "Stage all changes before committing (like git commit -a)")
	provider := flag.String("provider", "", "Override provider (free, claude, openai, ollama, openrouter)")
	model := flag.String("model", "", "Override model for the selected provider")
	listModels := flag.Bool("list-models", false, "Show available providers and built-in free models")
	ver := flag.Bool("version", false, "Print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `sage %s — Git AI Commit

Usage:
  sage [flags]

Flags:
`, Version)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  sage                        generate and commit
  sage --dry-run              generate only, no commit
  sage --provider free        use built-in free models
  sage --provider free --model qwen/qwen3-coder:free
  sage --provider ollama      use local Ollama model
  sage --list-models          show selectable models

Config file: ~/.config/sage/config.json
Docs:        github.com/vahidid/sage
`)
	}

	flag.Parse()

	if *ver {
		fmt.Println("sage", Version)
		return
	}

	if *listModels {
		printModels()
		return
	}

	if err := run(*dryRun, *stageAll, *provider, *model); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(dryRun bool, stageAll bool, providerOverride string, modelOverride string) error {
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
	if modelOverride != "" {
		applyModelOverride(cfg, modelOverride)
	}

	// 3. stage all if requested
	if stageAll {
		if git.HasUnstagedChanges() {
			fmt.Println("📦 Staging all changes...")
			if err := git.StageAll(); err != nil {
				return err
			}
		}
	}

	// 4. staged diff
	fmt.Println("📋 Reading staged changes...")
	diff, err := git.GetStagedDiff()
	if err != nil {
		return err
	}

	// 5. resolve provider
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
	case "free":
		apiKey := freeOpenRouterAPIKey()
		if apiKey == "" {
			return nil, fmt.Errorf(
				"❌ built-in free models are not enabled in this build\n" +
					"   Install an official release binary, or build with:\n" +
					"   go build -ldflags=\"-X main.BuiltinOpenRouterAPIKey=$SAGE_FREE_OPENROUTER_API_KEY\" .\n" +
					"   For local development, you can also set SAGE_FREE_OPENROUTER_API_KEY.",
			)
		}
		return ai.NewOpenRouterProvider(apiKey, cfg.Free.Model), nil

	case "claude":
		if cfg.Claude.APIKey == "" {
			return nil, fmt.Errorf(
				"❌ Claude API key not set\n"+
					"   Set env var:  export ANTHROPIC_API_KEY=sk-ant-...\n"+
					"   Or add it to: %s", config.FilePath(),
			)
		}
		return ai.NewClaudeProvider(cfg.Claude.APIKey, cfg.Claude.Model), nil

	case "openai":
		if cfg.OpenAI.APIKey == "" {
			return nil, fmt.Errorf(
				"❌ OpenAI API key not set\n"+
					"   Set env var:  export OPENAI_API_KEY=sk-...\n"+
					"   Or add it to: %s", config.FilePath(),
			)
		}
		return ai.NewOpenAIProvider(cfg.OpenAI.APIKey, cfg.OpenAI.Model), nil

	case "ollama":
		return ai.NewOllamaProvider(cfg.Ollama.Host, cfg.Ollama.Model), nil

	case "openrouter":
		if cfg.OpenRouter.APIKey == "" {
			return nil, fmt.Errorf(
				"❌ OpenRouter API key not set\n"+
					"   Set env var:  export OPENROUTER_API_KEY=sk-or-...\n"+
					"   Or add it to: %s", config.FilePath(),
			)
		}
		return ai.NewOpenRouterProvider(cfg.OpenRouter.APIKey, cfg.OpenRouter.Model), nil

	default:
		return nil, fmt.Errorf("❌ unknown provider %q — choose: free, claude, openai, ollama, openrouter", cfg.Provider)
	}
}

func applyModelOverride(cfg *config.Config, model string) {
	switch cfg.Provider {
	case "free":
		if preset, err := config.FreeModelByChoice(model); err == nil {
			cfg.Free.Model = preset.ID
			return
		}
		cfg.Free.Model = model
	case "claude":
		cfg.Claude.Model = model
	case "openai":
		cfg.OpenAI.Model = model
	case "ollama":
		cfg.Ollama.Model = model
	case "openrouter":
		cfg.OpenRouter.Model = model
	}
}

func freeOpenRouterAPIKey() string {
	if BuiltinOpenRouterAPIKey != "" {
		return BuiltinOpenRouterAPIKey
	}
	return os.Getenv("SAGE_FREE_OPENROUTER_API_KEY")
}

func printModels() {
	fmt.Println("Providers:")
	fmt.Println("  free        built-in OpenRouter free models; no user API key")
	fmt.Println("  claude      Anthropic API key")
	fmt.Println("  openai      OpenAI API key")
	fmt.Println("  ollama      local model")
	fmt.Println("  openrouter  custom OpenRouter API key/model")
	fmt.Println()
	fmt.Println("Built-in free models:")
	for i, model := range config.FreeModels {
		fmt.Printf("  [%d] %-38s %s (%s)\n", i+1, model.ID, model.Name, model.Description)
	}
}

func confirm(prompt string) bool {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return input == "" || input == "y" || input == "Y" || input == "yes"
}
