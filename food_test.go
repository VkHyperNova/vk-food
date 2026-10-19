package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tests in this file (each function is a group of named subtests):
//   TestLoadFood      - reading food.json
//   TestSaveFood      - writing food.json
//   TestAskFloat      - asking for one number
//   TestAskFloatKeep  - asking for one number, Enter keeps the old value
//   TestAskFood       - asking all the questions for a new food
//   TestAskFoodKeep   - asking all the questions, Enter keeps the old values
//   TestIndexOfFood   - finding a food by name
//   TestCheckName     - empty and duplicate names
//   TestAddFood       - the "add-food" command
//   TestUpdateFood    - the "update-food" command
//   TestDeleteFood    - the "delete-food" command
//
// UndoFood is tested in log_test.go. This file also uses the helpers
// useTempData, withArgs, seedLogs, exists, readFile and quiet from log_test.go,
// and capture from report_test.go.
//
// Run one group:  go test -v -run TestUpdateFood
// Run one test (spaces in the name become _):
//                   go test -v -run 'TestUpdateFood/rename_changes_the_log'
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
// askFloat
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
	// askFloat would call os.Exit and stop the whole test run.
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
// askFloatKeep
// ============================================================

func TestAskFloatKeep(t *testing.T) {
	t.Run("Enter keeps the current value", func(t *testing.T) {
		quiet(t)
		reader := bufio.NewReader(strings.NewReader("\n"))

		if got := askFloatKeep(reader, "Protein (g)", 13); got != 13 {
			t.Errorf("got %v, want 13", got)
		}
	})

	t.Run("spaces only also keeps the current value", func(t *testing.T) {
		quiet(t)
		reader := bufio.NewReader(strings.NewReader("   \n"))

		if got := askFloatKeep(reader, "Protein (g)", 13); got != 13 {
			t.Errorf("got %v, want 13", got)
		}
	})

	t.Run("a number replaces the current value", func(t *testing.T) {
		quiet(t)
		reader := bufio.NewReader(strings.NewReader("14.5\n"))

		if got := askFloatKeep(reader, "Protein (g)", 13); got != 14.5 {
			t.Errorf("got %v, want 14.5", got)
		}
	})

	t.Run("zero replaces the current value", func(t *testing.T) {
		quiet(t)
		reader := bufio.NewReader(strings.NewReader("0\n"))

		if got := askFloatKeep(reader, "Protein (g)", 13); got != 0 {
			t.Errorf("got %v, want 0", got)
		}
	})

	// A bad answer is refused (it does NOT keep the old value), the next line is used.
	t.Run("asks again after a bad answer", func(t *testing.T) {
		bad := []string{"abc", "-5", "0,01", "NaN", "Inf", "12g"}
		for _, answer := range bad {
			t.Run("bad answer "+answer, func(t *testing.T) {
				quiet(t)
				reader := bufio.NewReader(strings.NewReader(answer + "\n3.5\n"))

				if got := askFloatKeep(reader, "Protein (g)", 13); got != 3.5 {
					t.Errorf("got %v, want 3.5 (the answer after the bad one)", got)
				}
			})
		}
	})

	t.Run("prompt shows the label and the current value in green", func(t *testing.T) {
		reader := bufio.NewReader(strings.NewReader("\n"))

		out := capture(t, func() { askFloatKeep(reader, "Protein (g)", 13.5) })

		want := "Protein (g) [" + Green + "13.5" + Reset + "]: "
		if out != want {
			t.Errorf("prompt = %q, want %q", out, want)
		}
	})
}

// ============================================================
// askFood
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
// askFoodKeep
// ============================================================

func TestAskFoodKeep(t *testing.T) {
	current := Food{Name: "Oats", Kcal: 370, Protein: 13, Carbs: 60, Fat: 7, Fiber: 10, Sugar: 1, Salt: 0.02}

	t.Run("Enter on every question keeps everything", func(t *testing.T) {
		withStdin(t, "\n\n\n\n\n\n\n\n") // name + 7 numbers

		if got := askFoodKeep(current); got != current {
			t.Errorf("got %+v, want %+v", got, current)
		}
	})

	t.Run("changes only the answered fields", func(t *testing.T) {
		// name, kcal, protein, carbs, fat, fiber, sugar, salt
		withStdin(t, "\n\n14\n\n\n\n\n0\n")

		want := current
		want.Protein = 14
		want.Salt = 0 // typing 0 sets zero, it does not keep the old value
		if got := askFoodKeep(current); got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("a new name replaces the old one", func(t *testing.T) {
		withStdin(t, "  Rolled oats  \n\n\n\n\n\n\n\n")

		want := current
		want.Name = "Rolled oats"
		if got := askFoodKeep(current); got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("asks again after a bad number", func(t *testing.T) {
		// "abc" is refused, so 400 is read for calories
		withStdin(t, "\nabc\n400\n\n\n\n\n\n\n") // name, bad, 400, then 6 more Enters

		want := current
		want.Kcal = 400
		if got := askFoodKeep(current); got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("prompt shows the current name in green", func(t *testing.T) {
		withStdin(t, "\n\n\n\n\n\n\n\n")

		out := capture(t, func() { askFoodKeep(current) })

		if !strings.HasPrefix(out, "Food name ["+Green+"Oats"+Reset+"]: ") {
			t.Errorf("unexpected first prompt: %q", out)
		}
	})
}

// ============================================================
// indexOfFood
// ============================================================

func TestIndexOfFood(t *testing.T) {
	foods := []Food{{Name: "Oats"}, {Name: "Cheese"}, {Name: "Peanut butter"}}

	t.Run("finds the position", func(t *testing.T) {
		if got := indexOfFood(foods, "Peanut butter"); got != 2 {
			t.Errorf("got %d, want 2", got)
		}
	})

	t.Run("ignores upper and lower case", func(t *testing.T) {
		if got := indexOfFood(foods, "cHeEsE"); got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("unknown food gives -1", func(t *testing.T) {
		if got := indexOfFood(foods, "Rice"); got != -1 {
			t.Errorf("got %d, want -1", got)
		}
	})

	t.Run("partial name does not match", func(t *testing.T) {
		if got := indexOfFood(foods, "Peanut"); got != -1 {
			t.Errorf("got %d, want -1", got)
		}
	})

	t.Run("empty list gives -1", func(t *testing.T) {
		if got := indexOfFood(nil, "Oats"); got != -1 {
			t.Errorf("got %d, want -1", got)
		}
	})
}

// ============================================================
// checkName
// ============================================================

func TestCheckName(t *testing.T) {
	foods := []Food{{Name: "Oats"}, {Name: "Cheese"}}

	t.Run("new name is fine", func(t *testing.T) {
		if err := checkName(foods, "Rice", -1); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("empty name is an error", func(t *testing.T) {
		if err := checkName(foods, "", -1); err == nil {
			t.Error("expected an error")
		}
	})

	t.Run("existing name is an error, any case", func(t *testing.T) {
		for _, name := range []string{"Oats", "oats", "OATS"} {
			if err := checkName(foods, name, -1); err == nil {
				t.Errorf("%q should be refused", name)
			}
		}
	})

	t.Run("a food may keep its own name", func(t *testing.T) {
		// skip = 0 is the position of Oats, the food being updated
		if err := checkName(foods, "Oats", 0); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if err := checkName(foods, "OATS", 0); err != nil {
			t.Errorf("changing only the case should be fine: %v", err)
		}
	})

	t.Run("a food may not take the name of another food", func(t *testing.T) {
		if err := checkName(foods, "Cheese", 0); err == nil {
			t.Error("expected an error")
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

	t.Run("existing name is refused, any case", func(t *testing.T) {
		for _, name := range []string{"Oats", "oats"} {
			useTempData(t)
			before := readFile(t, FoodFile)
			withStdin(t, name+"\n370\n13\n60\n7\n10\n1\n0.02\n")

			if err := AddFood(); err == nil {
				t.Errorf("%q should be refused", name)
			}
			if got := readFile(t, FoodFile); got != before {
				t.Errorf("food file changed after refused name %q", name)
			}
		}
	})

	t.Run("empty name is refused", func(t *testing.T) {
		useTempData(t)
		before := readFile(t, FoodFile)
		withStdin(t, "\n370\n13\n60\n7\n10\n1\n0.02\n")

		if err := AddFood(); err == nil {
			t.Error("expected an error for an empty name")
		}
		if got := readFile(t, FoodFile); got != before {
			t.Error("food file changed after refused name")
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
// UpdateFood
// ============================================================

func TestUpdateFood(t *testing.T) {
	// 8 empty answers: name + 7 numbers, all kept
	allEnter := "\n\n\n\n\n\n\n\n"

	t.Run("Enter everywhere changes nothing", func(t *testing.T) {
		useTempData(t)
		seedLogs(t)
		foodsBefore, logsBefore := readFile(t, FoodFile), readFile(t, LogFile)
		withArgs(t, "update-food", "Cheese")
		withStdin(t, allEnter)

		if err := UpdateFood(); err != nil {
			t.Fatal(err)
		}
		if got := readFile(t, FoodFile); got != foodsBefore {
			t.Errorf("food file changed:\n%s", got)
		}
		if got := readFile(t, LogFile); got != logsBefore {
			t.Errorf("log file changed:\n%s", got)
		}
	})

	t.Run("changes the values and leaves the other foods alone", func(t *testing.T) {
		useTempData(t)
		withArgs(t, "update-food", "Cheese")
		withStdin(t, "\n450\n\n\n\n\n\n1.9\n") // only Kcal and Salt

		if err := UpdateFood(); err != nil {
			t.Fatal(err)
		}

		foods, _ := LoadFood(FoodFile)
		if len(foods) != 3 {
			t.Fatalf("got %d foods, want 3", len(foods))
		}
		if foods[1].Name != "Cheese" || foods[1].Kcal != 450 || foods[1].Salt != 1.9 {
			t.Errorf("Cheese not updated: %+v", foods[1])
		}
		if foods[0] != (Food{Name: "Oats", Kcal: 370}) || foods[2] != (Food{Name: "Peanut butter", Kcal: 600}) {
			t.Errorf("other foods changed: %+v", foods)
		}
	})

	t.Run("food is found in any case", func(t *testing.T) {
		useTempData(t)
		withArgs(t, "update-food", "cheese")
		withStdin(t, "\n450\n\n\n\n\n\n\n")

		if err := UpdateFood(); err != nil {
			t.Fatal(err)
		}
		foods, _ := LoadFood(FoodFile)
		if foods[1].Kcal != 450 {
			t.Errorf("Cheese not updated: %+v", foods[1])
		}
	})

	t.Run("rename changes the log", func(t *testing.T) {
		useTempData(t)
		seedLogs(t)
		withArgs(t, "update-food", "Oats")
		withStdin(t, "Rolled oats\n\n\n\n\n\n\n\n")

		if err := UpdateFood(); err != nil {
			t.Fatal(err)
		}

		foods, _ := LoadFood(FoodFile)
		if foods[0].Name != "Rolled oats" || foods[0].Kcal != 370 {
			t.Errorf("food not renamed or values lost: %+v", foods[0])
		}

		logs, _ := LoadLog(LogFile)
		if logs[0].Food != "Rolled oats" || logs[0].Grams != 50 {
			t.Errorf("log entry not renamed: %+v", logs[0])
		}
		if logs[1].Food != "Cheese" || logs[2].Food != "Peanut butter" {
			t.Errorf("other log entries changed: %+v", logs)
		}
	})

	t.Run("changing only the case of the name is allowed", func(t *testing.T) {
		useTempData(t)
		withArgs(t, "update-food", "Oats")
		withStdin(t, "OATS\n\n\n\n\n\n\n\n")

		if err := UpdateFood(); err != nil {
			t.Fatal(err)
		}
		foods, _ := LoadFood(FoodFile)
		if foods[0].Name != "OATS" {
			t.Errorf("Name = %q, want OATS", foods[0].Name)
		}
	})

	t.Run("name of another food is refused and nothing changes", func(t *testing.T) {
		useTempData(t)
		seedLogs(t)
		foodsBefore, logsBefore := readFile(t, FoodFile), readFile(t, LogFile)
		withArgs(t, "update-food", "Oats")
		withStdin(t, "cheese\n999\n\n\n\n\n\n\n")

		if err := UpdateFood(); err == nil {
			t.Fatal("expected an error")
		}
		if got := readFile(t, FoodFile); got != foodsBefore {
			t.Error("food file changed")
		}
		if got := readFile(t, LogFile); got != logsBefore {
			t.Error("log file changed")
		}
	})

	t.Run("unknown food is an error", func(t *testing.T) {
		useTempData(t)
		withArgs(t, "update-food", "Rice")

		if err := UpdateFood(); err == nil {
			t.Error("expected an error")
		}
	})

	t.Run("missing food name is an error", func(t *testing.T) {
		useTempData(t)
		withArgs(t, "update-food")

		if err := UpdateFood(); err == nil {
			t.Error("expected an error")
		}
	})

	t.Run("broken food file is not overwritten", func(t *testing.T) {
		useTempData(t)
		os.WriteFile(FoodFile, []byte("not json"), 0o644)
		withArgs(t, "update-food", "Oats")

		if err := UpdateFood(); err == nil {
			t.Fatal("expected an error")
		}
		if got := readFile(t, FoodFile); got != "not json" {
			t.Errorf("broken file was overwritten with: %s", got)
		}
	})

	t.Run("broken log file stops the update before any change", func(t *testing.T) {
		useTempData(t)
		foodsBefore := readFile(t, FoodFile)
		os.WriteFile(LogFile, []byte("not json"), 0o644)
		withArgs(t, "update-food", "Oats")
		withStdin(t, "Rolled oats\n\n\n\n\n\n\n\n")

		if err := UpdateFood(); err == nil {
			t.Fatal("expected an error")
		}
		if got := readFile(t, FoodFile); got != foodsBefore {
			t.Error("food file changed even though the log could not be loaded")
		}
		if got := readFile(t, LogFile); got != "not json" {
			t.Errorf("broken log was overwritten with: %s", got)
		}
	})
}

// ============================================================
// DeleteFood
// ============================================================

func TestDeleteFood(t *testing.T) {
	names := func(foods []Food) []string {
		var out []string
		for _, f := range foods {
			out = append(out, f.Name)
		}
		return out
	}

	t.Run("removes the food and keeps the order of the others", func(t *testing.T) {
		useTempData(t) // Oats, Cheese, Peanut butter
		withArgs(t, "delete-food", "Cheese")

		if err := DeleteFood(); err != nil {
			t.Fatal(err)
		}

		foods, _ := LoadFood(FoodFile)
		if got := strings.Join(names(foods), ","); got != "Oats,Peanut butter" {
			t.Errorf("got %s, want Oats,Peanut butter", got)
		}
	})

	t.Run("works on the first and the last food", func(t *testing.T) {
		for arg, want := range map[string]string{"Oats": "Cheese,Peanut butter", "Peanut butter": "Oats,Cheese"} {
			useTempData(t)
			withArgs(t, "delete-food", arg)

			if err := DeleteFood(); err != nil {
				t.Fatal(err)
			}
			foods, _ := LoadFood(FoodFile)
			if got := strings.Join(names(foods), ","); got != want {
				t.Errorf("after deleting %s: got %s, want %s", arg, got, want)
			}
		}
	})

	t.Run("food is found in any case", func(t *testing.T) {
		useTempData(t)
		withArgs(t, "delete-food", "cHeEsE")

		if err := DeleteFood(); err != nil {
			t.Fatal(err)
		}
		foods, _ := LoadFood(FoodFile)
		if len(foods) != 2 {
			t.Errorf("got %d foods, want 2", len(foods))
		}
	})

	t.Run("deleting the only food leaves an empty list", func(t *testing.T) {
		useTempData(t)
		writeJSON(FoodFile, []Food{{Name: "Oats"}})
		withArgs(t, "delete-food", "Oats")

		if err := DeleteFood(); err != nil {
			t.Fatal(err)
		}
		foods, err := LoadFood(FoodFile)
		if err != nil || len(foods) != 0 {
			t.Errorf("foods = %+v, err = %v, want an empty list", foods, err)
		}
	})

	t.Run("food used in the log is refused and nothing changes", func(t *testing.T) {
		useTempData(t)
		seedLogs(t) // Oats, Cheese, Peanut butter are all logged
		foodsBefore, logsBefore := readFile(t, FoodFile), readFile(t, LogFile)
		withArgs(t, "delete-food", "oats") // different case, still the same food

		if err := DeleteFood(); err == nil {
			t.Fatal("expected an error")
		}
		if got := readFile(t, FoodFile); got != foodsBefore {
			t.Error("food file changed")
		}
		if got := readFile(t, LogFile); got != logsBefore {
			t.Error("log file changed")
		}
	})

	t.Run("food that is not in the log can be deleted", func(t *testing.T) {
		useTempData(t)
		writeJSON(LogFile, []Log{{Food: "Oats", Grams: 50}}) // only Oats is logged
		withArgs(t, "delete-food", "Cheese")

		if err := DeleteFood(); err != nil {
			t.Fatal(err)
		}
		foods, _ := LoadFood(FoodFile)
		if len(foods) != 2 {
			t.Errorf("got %d foods, want 2", len(foods))
		}
	})

	t.Run("unknown food is an error", func(t *testing.T) {
		useTempData(t)
		before := readFile(t, FoodFile)
		withArgs(t, "delete-food", "Rice")

		if err := DeleteFood(); err == nil {
			t.Error("expected an error")
		}
		if got := readFile(t, FoodFile); got != before {
			t.Error("food file changed")
		}
	})

	t.Run("missing food name is an error", func(t *testing.T) {
		useTempData(t)
		withArgs(t, "delete-food")

		if err := DeleteFood(); err == nil {
			t.Error("expected an error")
		}
	})

	t.Run("broken log file stops the delete", func(t *testing.T) {
		useTempData(t)
		before := readFile(t, FoodFile)
		os.WriteFile(LogFile, []byte("not json"), 0o644)
		withArgs(t, "delete-food", "Cheese")

		if err := DeleteFood(); err == nil {
			t.Fatal("expected an error")
		}
		if got := readFile(t, FoodFile); got != before {
			t.Error("food file changed even though the log could not be loaded")
		}
	})
}

// ============================================================
// Helper (used by the tests above)
// ============================================================

// withStdin makes os.Stdin read from the given text for the length of the test,
// and hides the prompts the app prints.
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
