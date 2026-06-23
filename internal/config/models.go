package config

import "fmt"

type FreeModel struct {
	ID          string
	Name        string
	Description string
}

const (
	DefaultFreeLLMAPIModel = "auto"
)

var FreeModels = []FreeModel{
	{
		ID:          DefaultFreeLLMAPIModel,
		Name:        "FreeLLMApi Llama 3.3 70B",
		Description: "pinned instruction-following free model",
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
