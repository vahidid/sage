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
	model, err := FreeModelByChoice("2")
	if err != nil {
		t.Fatalf("FreeModelByChoice() error = %v", err)
	}
	if model.ID != FreeModels[1].ID {
		t.Fatalf("FreeModelByChoice() = %q, want %q", model.ID, FreeModels[1].ID)
	}
}

func TestFreeModelByChoiceAcceptsModelID(t *testing.T) {
	model, err := FreeModelByChoice(FreeModels[2].ID)
	if err != nil {
		t.Fatalf("FreeModelByChoice() error = %v", err)
	}
	if model.ID != FreeModels[2].ID {
		t.Fatalf("FreeModelByChoice() = %q, want %q", model.ID, FreeModels[2].ID)
	}
}

func TestFreeModelByChoiceRejectsUnknownChoice(t *testing.T) {
	if _, err := FreeModelByChoice("unknown"); err == nil {
		t.Fatal("FreeModelByChoice() error = nil, want non-nil")
	}
}
