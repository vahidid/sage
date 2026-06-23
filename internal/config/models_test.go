package config

import "testing"

func TestFreeModelByChoiceDefaultsToFirstModel(t *testing.T) {
	model, err := FreeModelByChoice("")
	if err != nil {
		t.Fatalf("FreeModelByChoice() error = %v", err)
	}
	if model.ID != DefaultFreeModel() {
		t.Fatalf("FreeModelByChoice() = %q, want %q", model.ID, DefaultFreeModel())
	}
}

func TestFreeModelByChoiceAcceptsNumber(t *testing.T) {
	model, err := FreeModelByChoice("1")
	if err != nil {
		t.Fatalf("FreeModelByChoice() error = %v", err)
	}
	if model.ID != FreeModels[0].ID {
		t.Fatalf("FreeModelByChoice() = %q, want %q", model.ID, FreeModels[0].ID)
	}
}

func TestFreeModelByChoiceAcceptsModelID(t *testing.T) {
	model, err := FreeModelByChoice(DefaultFreeLLMAPIModel)
	if err != nil {
		t.Fatalf("FreeModelByChoice() error = %v", err)
	}
	if model.ID != DefaultFreeLLMAPIModel {
		t.Fatalf("FreeModelByChoice() = %q, want %q", model.ID, DefaultFreeLLMAPIModel)
	}
}

func TestFreeModelByChoiceRejectsUnknownChoice(t *testing.T) {
	if _, err := FreeModelByChoice("unknown"); err == nil {
		t.Fatal("FreeModelByChoice() error = nil, want non-nil")
	}
}
