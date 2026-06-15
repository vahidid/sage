package config

import "fmt"

type FreeModel struct {
	ID          string
	Name        string
	Description string
}

var FreeModels = []FreeModel{
	{
		ID:          "qwen/qwen3-coder:free",
		Name:        "Qwen3 Coder",
		Description: "best default for code-heavy diffs",
	},
	{
		ID:          "openai/gpt-oss-20b:free",
		Name:        "GPT OSS 20B",
		Description: "fast general-purpose commit messages",
	},
	{
		ID:          "google/gemma-4-26b-a4b-it:free",
		Name:        "Gemma 4 26B",
		Description: "balanced open model",
	},
	{
		ID:          "meta-llama/llama-3.3-70b-instruct:free",
		Name:        "Llama 3.3 70B",
		Description: "strong general instruction following",
	},
}

func DefaultFreeModel() string {
	return FreeModels[0].ID
}

func FreeModelByChoice(choice string) (FreeModel, error) {
	if choice == "" {
		return FreeModels[0], nil
	}
	for i, model := range FreeModels {
		if choice == fmt.Sprintf("%d", i+1) || choice == model.ID {
			return model, nil
		}
	}
	return FreeModel{}, fmt.Errorf("invalid free model choice %q", choice)
}
