package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// useTempFoodFile points FoodFile at a temp folder and fakes the USB drive,
// so tests never touch your real data or ask "Mount?".
func useTempFoodFile(t *testing.T) string {
	t.Helper()

	oldFile, oldBackup := FoodFile, FoodFileBackup
	oldCheck, oldAsk := checkMounted, askMount
	t.Cleanup(func() {
		FoodFile, FoodFileBackup = oldFile, oldBackup
		checkMounted, askMount = oldCheck, oldAsk
	})

	dir := t.TempDir()
	FoodFile = filepath.Join(dir, "local", "food.json")
	FoodFileBackup = filepath.Join(dir, "usb", "food.json")
	checkMounted = func() (bool, error) { return false, nil }
	askMount = func() (bool, error) { return false, nil } // user answers No

	return FoodFile
}

// withStdin makes os.Stdin read from the given text for the length of the test.
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

	old := os.Stdin
	os.Stdin = f
	t.Cleanup(func() {
		os.Stdin = old
		f.Close()
	})
}

func TestLoadFoodMissingFile(t *testing.T) {
	foods, err := LoadFood(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatalf("missing file should not be an error, got %v", err)
	}
	if len(foods) != 0 {
		t.Errorf("got %d foods, want 0", len(foods))
	}
}

func TestLoadFoodValid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "food.json")
	data := `[{"Name":"Oats","Kcal":370,"Protein":13,"Carbs":60,"Fat":7,"Fiber":10,"Sugar":1,"Salt":0.02}]`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	foods, err := LoadFood(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(foods) != 1 || foods[0].Name != "Oats" || foods[0].Salt != 0.02 {
		t.Errorf("unexpected foods: %+v", foods)
	}
}

func TestLoadFoodBadJSON(t *testing.T) {
	for _, content := range []string{"not json", `{}`, `{"foods": []}`} {
		path := filepath.Join(t.TempDir(), "food.json")
		os.WriteFile(path, []byte(content), 0o644)

		if _, err := LoadFood(path); err == nil {
			t.Errorf("content %q should give an error", content)
		}
	}
}

func TestAskFloatValid(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("  12.5  \n"))

	if got := AskFloat(reader, ""); got != 12.5 {
		t.Errorf("got %v, want 12.5", got)
	}
}

func TestAskFloatRetriesOnBadInput(t *testing.T) {
	// the last line is valid, otherwise AskFloat would hit EOF and call os.Exit
	input := "abc\n-5\n0,01\nNaN\nInf\n\n3.5\n"
	reader := bufio.NewReader(strings.NewReader(input))

	if got := AskFloat(reader, ""); got != 3.5 {
		t.Errorf("got %v, want 3.5 after the bad lines", got)
	}
}

func TestAskFloatAcceptsZero(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("0\n"))

	if got := AskFloat(reader, ""); got != 0 {
		t.Errorf("got %v, want 0", got)
	}
}

func TestAskFood(t *testing.T) {
	withStdin(t, "Oats\n370\n13\n60\n7\n10\n1\n0.02\n")

	want := Food{Name: "Oats", Kcal: 370, Protein: 13, Carbs: 60, Fat: 7, Fiber: 10, Sugar: 1, Salt: 0.02}
	if got := AskFood(); got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestAskFoodNameWithSpaces(t *testing.T) {
	withStdin(t, "  Peanut butter  \n600\n25\n20\n50\n6\n9\n0.01\n")

	got := AskFood()
	if got.Name != "Peanut butter" {
		t.Errorf("Name = %q, want %q", got.Name, "Peanut butter")
	}
	if got.Kcal != 600 || got.Salt != 0.01 {
		t.Errorf("values shifted: %+v", got)
	}
}

func TestAddFoodFirstRun(t *testing.T) {
	path := useTempFoodFile(t)
	withStdin(t, "Oats\n370\n13\n60\n7\n10\n1\n0.02\n")

	if err := AddFood(); err != nil {
		t.Fatal(err)
	}

	foods, err := LoadFood(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(foods) != 1 || foods[0].Name != "Oats" {
		t.Errorf("unexpected foods: %+v", foods)
	}
}

func TestAddFoodAppends(t *testing.T) {
	path := useTempFoodFile(t)

	withStdin(t, "Oats\n370\n13\n60\n7\n10\n1\n0.02\n")
	if err := AddFood(); err != nil {
		t.Fatal(err)
	}
	withStdin(t, "Cheese\n400\n25\n1\n33\n0\n1\n1.8\n")
	if err := AddFood(); err != nil {
		t.Fatal(err)
	}

	foods, err := LoadFood(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(foods) != 2 || foods[0].Name != "Oats" || foods[1].Name != "Cheese" {
		t.Errorf("old food lost or order wrong: %+v", foods)
	}
}

func TestAddFoodBrokenFileIsNotOverwritten(t *testing.T) {
	path := useTempFoodFile(t)
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte("not json"), 0o644)
	withStdin(t, "Oats\n370\n13\n60\n7\n10\n1\n0.02\n")

	if err := AddFood(); err == nil {
		t.Fatal("expected an error for the broken file")
	}

	data, _ := os.ReadFile(path)
	if string(data) != "not json" {
		t.Errorf("broken file was overwritten with: %s", data)
	}
}

func TestSaveFoodRoundtrip(t *testing.T) {
	path := useTempFoodFile(t)
	in := []Food{{Name: "Oats", Kcal: 370}, {Name: "Cheese", Kcal: 400, Salt: 1.8}}

	if err := SaveFood(in); err != nil {
		t.Fatal(err)
	}

	out, err := LoadFood(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[1] != in[1] {
		t.Errorf("got %+v, want %+v", out, in)
	}
}
