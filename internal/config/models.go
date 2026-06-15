package config

import "fmt"

type FreeModel struct {
	ID          string
	Name        string
	Description string
}

var FreeModels = []FreeModel{
	{
		ID:          "nvidia/nemotron-3-ultra-550b-a55b:free",
		Name:        "NVIDIA Nemotron 3 Ultra",
		Description: "large reasoning model for complex diffs",
	},
	{
		ID:          "nex-agi/nex-n2-pro:free",
		Name:        "Nex AGI Nex-N2-Pro",
		Description: "agentic MoE model with strong code reasoning",
	},
	{
		ID:          "google/gemma-4-31b-it:free",
		Name:        "Google Gemma 4 31B",
		Description: "long-context instruct model for larger changes",
	},
	{
		ID:          "liquid/lfm-2.5-1.2b-instruct:free",
		Name:        "LiquidAI LFM2.5 1.2B",
		Description: "compact low-latency model for quick commits",
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
