package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "foods.json")
	want := []Food{
		{Name: "Oats", Kcal: 370, Protein: 13, Carbs: 60, Fat: 7},
		{Name: "Banana", Kcal: 99, Carbs: 23},
	}

	if err := writeJSON(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFood(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != len(want) {
		t.Fatalf("got %d foods, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("food %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := LoadFood(filepath.Join(t.TempDir(), "nope.json"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected ErrNotExist, got %v", err)
	}
}

func TestLoadBadJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "foods.json")
	os.WriteFile(path, []byte("garbage"), 0o644)

	if _, err := LoadFood(path); err == nil {
		t.Error("expected an error for invalid JSON, got nil")
	}
}
