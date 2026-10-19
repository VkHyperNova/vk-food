package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tests in this file (each function is a group of named subtests):
//   TestLoadFood  - reading food.json
//   TestSaveFood  - writing food.json
//   TestAskFloat  - asking for one number
//   TestAskFood   - asking all the questions for a new food
//   TestAddFood   - the "add-food" command
//
// UndoFood is tested in log_test.go. This file also uses the helpers
// useTempData, exists, readFile and quiet from log_test.go.
//
// Run one group:  go test -v -run TestAddFood
// Run one test (spaces in the name become _):
//                   go test -v -run 'TestAddFood/first_run'
//
// The helper function is at the bottom of the file.

// ============================================================
// LoadFood
// ============================================================

func TestLoadFood(t *testing.T) {
	t.Run("missing file is an empty list", func(t *testing.T) {
		foods, err := LoadFood(filepath.Join(t.TempDir(), "nope.json"))
		if err != nil {
			t.Fatalf("missing file should not be an error, got %v", err)
		}
		if len(foods) != 0 {
			t.Errorf("got %d foods, want 0", len(foods))
		}
	})

	t.Run("loads all fields", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "food.json")
		data := `[{"Name":"Oats","Kcal":370,"Protein":13,"Carbs":60,"Fat":7,"Fiber":10,"Sugar":1,"Salt":0.02}]`
		os.WriteFile(path, []byte(data), 0o644)

		foods, err := LoadFood(path)
		if err != nil {
			t.Fatal(err)
		}
		want := Food{Name: "Oats", Kcal: 370, Protein: 13, Carbs: 60, Fat: 7, Fiber: 10, Sugar: 1, Salt: 0.02}
		if len(foods) != 1 || foods[0] != want {
			t.Errorf("got %+v, want [%+v]", foods, want)
		}
	})

	t.Run("bad JSON is an error", func(t *testing.T) {
		for _, content := range []string{"not json", `{}`, `{"foods": []}`} {
			path := filepath.Join(t.TempDir(), "food.json")
			os.WriteFile(path, []byte(content), 0o644)

			if _, err := LoadFood(path); err == nil {
				t.Errorf("content %q should give an error", content)
			}
		}
	})
}

// ============================================================
// SaveFood
// ============================================================

func TestSaveFood(t *testing.T) {
	t.Run("save then load gives the same data", func(t *testing.T) {
		useTempData(t)
		in := []Food{{Name: "Oats", Kcal: 370}, {Name: "Cheese", Kcal: 400, Salt: 1.8}}

		if err := SaveFood(in); err != nil {
			t.Fatal(err)
		}

		out, err := LoadFood(FoodFile)
		if err != nil {
			t.Fatal(err)
		}
		if len(out) != 2 || out[0] != in[0] || out[1] != in[1] {
			t.Errorf("got %+v, want %+v", out, in)
		}
	})
}

// ============================================================
// AskFloat
// ============================================================

func TestAskFloat(t *testing.T) {
	t.Run("reads a number", func(t *testing.T) {
		quiet(t)
		reader := bufio.NewReader(strings.NewReader("12.5\n"))

		if got := askFloat(reader, ""); got != 12.5 {
			t.Errorf("got %v, want 12.5", got)
		}
	})

	t.Run("ignores spaces around the number", func(t *testing.T) {
		quiet(t)
		reader := bufio.NewReader(strings.NewReader("  12.5  \n"))

		if got := askFloat(reader, ""); got != 12.5 {
			t.Errorf("got %v, want 12.5", got)
		}
	})

	t.Run("accepts zero", func(t *testing.T) {
		quiet(t)
		reader := bufio.NewReader(strings.NewReader("0\n"))

		if got := askFloat(reader, ""); got != 0 {
			t.Errorf("got %v, want 0", got)
		}
	})

	// Each bad answer must be refused, and the next line (3.5) is used instead.
	// The input always ends with a valid number: at the end of the input
	// AskFloat would call os.Exit and stop the whole test run.
	t.Run("asks again after a bad answer", func(t *testing.T) {
		bad := []string{"abc", "-5", "0,01", "NaN", "Inf", "", "12g"}
		for _, answer := range bad {
			t.Run("bad answer "+answer, func(t *testing.T) {
				quiet(t)
				reader := bufio.NewReader(strings.NewReader(answer + "\n3.5\n"))

				if got := askFloat(reader, ""); got != 3.5 {
					t.Errorf("got %v, want 3.5 (the answer after the bad one)", got)
				}
			})
		}
	})
}

// ============================================================
// AskFood
// ============================================================

func TestAskFood(t *testing.T) {
	t.Run("reads all fields in order", func(t *testing.T) {
		withStdin(t, "Oats\n370\n13\n60\n7\n10\n1\n0.02\n")

		want := Food{Name: "Oats", Kcal: 370, Protein: 13, Carbs: 60, Fat: 7, Fiber: 10, Sugar: 1, Salt: 0.02}
		if got := askFood(); got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("name with spaces keeps all the words", func(t *testing.T) {
		withStdin(t, "  Peanut butter  \n600\n25\n20\n50\n6\n9\n0.01\n")

		got := askFood()
		if got.Name != "Peanut butter" {
			t.Errorf("Name = %q, want %q", got.Name, "Peanut butter")
		}
		if got.Kcal != 600 || got.Salt != 0.01 {
			t.Errorf("values shifted: %+v", got)
		}
	})

	t.Run("asks again after a bad number", func(t *testing.T) {
		// "abc" is refused, so 370 is read for calories
		withStdin(t, "Oats\nabc\n370\n13\n60\n7\n10\n1\n0.02\n")

		if got := askFood(); got.Kcal != 370 || got.Protein != 13 {
			t.Errorf("got %+v, want Kcal 370 and Protein 13", got)
		}
	})
}

// ============================================================
// AddFood
// ============================================================

func TestAddFood(t *testing.T) {
	t.Run("first run creates the food file", func(t *testing.T) {
		useTempData(t)
		os.Remove(FoodFile) // no foods added yet
		withStdin(t, "Oats\n370\n13\n60\n7\n10\n1\n0.02\n")

		if err := AddFood(); err != nil {
			t.Fatal(err)
		}

		foods, err := LoadFood(FoodFile)
		if err != nil {
			t.Fatal(err)
		}
		if len(foods) != 1 || foods[0].Name != "Oats" {
			t.Errorf("unexpected foods: %+v", foods)
		}
	})

	t.Run("appends to existing foods", func(t *testing.T) {
		useTempData(t) // already has Oats, Cheese, Peanut butter
		withStdin(t, "Rice\n130\n2.7\n28\n0.3\n0.4\n0\n0.01\n")

		if err := AddFood(); err != nil {
			t.Fatal(err)
		}

		foods, err := LoadFood(FoodFile)
		if err != nil {
			t.Fatal(err)
		}
		if len(foods) != 4 || foods[0].Name != "Oats" || foods[3].Name != "Rice" {
			t.Errorf("old foods lost or new one not last: %+v", foods)
		}
	})

	t.Run("broken food file is not overwritten", func(t *testing.T) {
		useTempData(t)
		os.WriteFile(FoodFile, []byte("not json"), 0o644)
		withStdin(t, "Oats\n370\n13\n60\n7\n10\n1\n0.02\n")

		if err := AddFood(); err == nil {
			t.Fatal("expected an error for the broken file")
		}
		if got := readFile(t, FoodFile); got != "not json" {
			t.Errorf("broken file was overwritten with: %s", got)
		}
	})
}

// ============================================================
// Helper (used by the tests above)
// ============================================================

// withStdin makes os.Stdin read from the given text for the length of the test,
// and hides the prompts AskFood prints.
func withStdin(t *testing.T, input string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "stdin.txt")
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	quiet(t)

	old := os.Stdin
	os.Stdin = f
	t.Cleanup(func() {
		os.Stdin = old
		f.Close()
	})
}
