package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAskFloat checks valid numeric input.
func TestAskFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"Integer", "100\n", 100},
		{"Decimal", "23.5\n", 23.5},
		{"Zero", "0\n", 0},
		{"Whitespace", "  12.5  \n", 12.5},
		{"Invalid then valid", "abc\n25.5\n", 25.5},
		{"Negative then valid", "-10\n15\n", 15},
		{"Empty then valid", "\n20\n", 20},
		{"Infinity then valid", "Inf\n10\n", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))

			result := AskFloat(reader, "Enter number: ")

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestAskFood checks that food information is read correctly.
func TestAskFood(t *testing.T) {
	input := strings.Join([]string{
		"Chicken Breast",
		"107",
		"23.5",
		"0",
		"2",
		"0",
		"0",
		"0.1",
		"",
	}, "\n")

	// Temporarily replace standard input.
	oldStdin := os.Stdin
	defer func() {
		os.Stdin = oldStdin
	}()

	tmpFile, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(input); err != nil {
		t.Fatal(err)
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	os.Stdin = tmpFile

	result := AskFood()

	expected := Food{
		Name:    "Chicken Breast",
		Kcal:    107,
		Protein: 23.5,
		Carbs:   0,
		Fat:     2,
		Fiber:   0,
		Sugar:   0,
		Salt:    0.1,
	}

	if result != expected {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

// TestLoadFood checks loading valid JSON data.
func TestLoadFood(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "foods.json")

	data := `[
		{
			"Name": "Chicken Breast",
			"Kcal": 107,
			"Protein": 23.5,
			"Carbs": 0,
			"Fat": 2,
			"Fiber": 0,
			"Sugar": 0,
			"Salt": 0.1
		}
	]`

	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}

	foods, err := LoadFood(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(foods) != 1 {
		t.Fatalf("Expected 1 food, got %d", len(foods))
	}

	if foods[0].Name != "Chicken Breast" {
		t.Errorf("Unexpected food name: %s", foods[0].Name)
	}

	if foods[0].Kcal != 107 {
		t.Errorf("Expected 107 kcal, got %v", foods[0].Kcal)
	}
}

// TestLoadFoodMissingFile checks first-run behavior.
func TestLoadFoodMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")

	foods, err := LoadFood(path)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if foods != nil {
		t.Errorf("Expected nil foods, got %+v", foods)
	}
}

// TestLoadFoodInvalidJSON checks malformed JSON.
func TestLoadFoodInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.json")

	if err := os.WriteFile(path, []byte("{invalid json"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFood(path)

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// TestLoadFoodEmptyArray checks an empty food database.
func TestLoadFoodEmptyArray(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.json")

	if err := os.WriteFile(path, []byte("[]"), 0600); err != nil {
		t.Fatal(err)
	}

	foods, err := LoadFood(path)

	if err != nil {
		t.Fatal(err)
	}

	if len(foods) != 0 {
		t.Errorf("Expected 0 foods, got %d", len(foods))
	}
}
