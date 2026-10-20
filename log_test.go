package main

import (
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Tests in this file (each function is a group of named subtests):
//   TestLoadLog         - reading log.json
//   TestSaveLog         - writing log.json
//   TestAddLog          - the "log <food> <grams>" command line (arguments)
//   TestLogFood         - adding one log entry (used by the command line and the picker)
//   TestLogInteractive  - the "log" picker, only the parts that work without a terminal
//   TestUndoLog         - the "undo" command (removes the last log entry)
//   TestUndoFood        - removes the last food (lives in food.go, tested here
//                         because it needs the same helpers)
//
// The picker form itself (arrow keys, typing grams) needs a real terminal,
// so it can only be tested by hand: go run . log
//
// Run one group:  go test -v -run TestLogFood
// Run one test (spaces in the name become _):
//                   go test -v -run 'TestLogFood/rejects_a_food'
//
// The helper functions are at the bottom of the file.

// ============================================================
// LoadLog
// ============================================================

func TestLoadLog(t *testing.T) {
	t.Run("missing file is an empty list", func(t *testing.T) {
		logs, err := LoadLog(filepath.Join(t.TempDir(), "nope.json"))
		if err != nil {
			t.Fatalf("missing file should not be an error, got %v", err)
		}
		if len(logs) != 0 {
			t.Errorf("got %d entries, want 0", len(logs))
		}
	})

	t.Run("bad JSON is an error", func(t *testing.T) {
		for _, content := range []string{"not json", `{}`} {
			path := filepath.Join(t.TempDir(), "log.json")
			os.WriteFile(path, []byte(content), 0o644)

			if _, err := LoadLog(path); err == nil {
				t.Errorf("content %q should give an error", content)
			}
		}
	})
}

// ============================================================
// SaveLog
// ============================================================

func TestSaveLog(t *testing.T) {
	t.Run("save then load gives the same data", func(t *testing.T) {
		path := useTempData(t)
		when := time.Date(2026, 10, 14, 12, 30, 0, 0, time.UTC)
		in := []Log{{Time: when, Food: "Oats", Grams: 50}, {Time: when, Food: "Cheese", Grams: 30.5}}

		if err := SaveLog(in); err != nil {
			t.Fatal(err)
		}

		out, err := LoadLog(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(out) != 2 || !out[0].Time.Equal(when) || out[1].Food != "Cheese" || out[1].Grams != 30.5 {
			t.Errorf("got %+v, want %+v", out, in)
		}
	})
}

// ============================================================
// AddLog (the command line: log <food> <grams>)
// ============================================================

// "go run . log" with no arguments opens the picker, which needs a terminal,
// so it is not tested here. The tests here always give at least a food name.
func TestAddLog(t *testing.T) {
	t.Run("logs the food and grams from the command line", func(t *testing.T) {
		path := useTempData(t)
		withArgs(t, "log", "Oats", "50")

		if err := AddLog(); err != nil {
			t.Fatal(err)
		}

		logs, err := LoadLog(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(logs) != 1 || logs[0].Food != "Oats" || logs[0].Grams != 50 {
			t.Errorf("unexpected logs: %+v", logs)
		}
	})

	t.Run("decimal grams", func(t *testing.T) {
		path := useTempData(t)
		withArgs(t, "log", "Oats", "12.5")

		if err := AddLog(); err != nil {
			t.Fatal(err)
		}

		logs, _ := LoadLog(path)
		if len(logs) != 1 || logs[0].Grams != 12.5 {
			t.Errorf("unexpected logs: %+v", logs)
		}
	})

	t.Run("bad arguments", func(t *testing.T) {
		tests := []struct {
			name string
			args []string
		}{
			{"missing grams", []string{"log", "Oats"}},
			{"grams not a number", []string{"log", "Oats", "abc"}},
			{"grams with unit", []string{"log", "Oats", "100g"}},
			{"grams with comma", []string{"log", "Oats", "1,5"}},
			{"zero grams", []string{"log", "Oats", "0"}},
			{"negative grams", []string{"log", "Oats", "-5"}},
			{"NaN grams", []string{"log", "Oats", "NaN"}},
			{"infinite grams", []string{"log", "Oats", "Inf"}},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				path := useTempData(t)
				withArgs(t, tc.args...)

				if err := AddLog(); err == nil {
					t.Error("expected an error")
				}
				if exists(path) {
					t.Error("nothing should be saved after bad input")
				}
			})
		}
	})
}

// ============================================================
// logFood
// ============================================================

func TestLogFood(t *testing.T) {
	t.Run("first run creates the log", func(t *testing.T) {
		path := useTempData(t)

		before := time.Now()
		if err := logFood("Oats", 50); err != nil {
			t.Fatal(err)
		}

		logs, err := LoadLog(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(logs) != 1 || logs[0].Food != "Oats" || logs[0].Grams != 50 {
			t.Fatalf("unexpected logs: %+v", logs)
		}
		if logs[0].Time.Before(before.Add(-time.Second)) || logs[0].Time.After(time.Now().Add(time.Second)) {
			t.Errorf("Time = %v, want about now", logs[0].Time)
		}
	})

	t.Run("appends to existing entries", func(t *testing.T) {
		path := useTempData(t)

		if err := logFood("Oats", 50); err != nil {
			t.Fatal(err)
		}
		if err := logFood("Cheese", 30); err != nil {
			t.Fatal(err)
		}

		logs, err := LoadLog(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(logs) != 2 || logs[0].Food != "Oats" || logs[1].Food != "Cheese" {
			t.Errorf("old entry lost or order wrong: %+v", logs)
		}
	})

	t.Run("stores the name from the food list", func(t *testing.T) {
		path := useTempData(t)

		if err := logFood("peanut butter", 30); err != nil { // typed in lower case
			t.Fatal(err)
		}

		logs, _ := LoadLog(path)
		if len(logs) != 1 || logs[0].Food != "Peanut butter" {
			t.Errorf("want the name from the food list (Peanut butter), got %+v", logs)
		}
	})

	t.Run("rejects a food that is not in the list", func(t *testing.T) {
		path := useTempData(t)

		if err := logFood("Pizza", 100); err == nil {
			t.Fatal("expected an error for a food that is not in the list")
		}
		if exists(path) {
			t.Error("nothing should be saved for an unknown food")
		}
	})

	t.Run("rejects when there are no foods at all", func(t *testing.T) {
		path := useTempData(t)
		os.Remove(FoodFile) // first run: no foods added yet

		if err := logFood("Oats", 50); err == nil {
			t.Fatal("expected an error when there are no foods at all")
		}
		if exists(path) {
			t.Error("nothing should be saved")
		}
	})

	t.Run("rejects bad grams", func(t *testing.T) {
		for name, grams := range map[string]float64{
			"zero":      0,
			"negative":  -5,
			"NaN":       math.NaN(),
			"infinite":  math.Inf(1),
			"-infinite": math.Inf(-1),
		} {
			t.Run(name, func(t *testing.T) {
				path := useTempData(t)

				if err := logFood("Oats", grams); err == nil {
					t.Error("expected an error")
				}
				if exists(path) {
					t.Error("nothing should be saved")
				}
			})
		}
	})

	t.Run("broken log file is not overwritten", func(t *testing.T) {
		path := useTempData(t)
		os.MkdirAll(filepath.Dir(path), 0o755)
		os.WriteFile(path, []byte("not json"), 0o644)

		if err := logFood("Oats", 50); err == nil {
			t.Fatal("expected an error for the broken file")
		}
		if got := readFile(t, path); got != "not json" {
			t.Errorf("broken file was overwritten with: %s", got)
		}
	})

	t.Run("broken food file is an error", func(t *testing.T) {
		path := useTempData(t)
		os.WriteFile(FoodFile, []byte("not json"), 0o644)

		if err := logFood("Oats", 50); err == nil {
			t.Fatal("expected an error for the broken food file")
		}
		if exists(path) {
			t.Error("nothing should be saved")
		}
	})

	t.Run("does not touch the food file", func(t *testing.T) {
		useTempData(t)
		before := readFile(t, FoodFile)

		if err := logFood("Oats", 50); err != nil {
			t.Fatal(err)
		}
		if got := readFile(t, FoodFile); got != before {
			t.Error("food file changed")
		}
	})
}

// ============================================================
// LogInteractive
// ============================================================

// Only the checks that happen BEFORE the form opens are tested here.
// If one of these tests ever hangs, the check was skipped and the form opened.
func TestLogInteractive(t *testing.T) {
	t.Run("no foods yet is an error", func(t *testing.T) {
		path := useTempData(t)
		os.Remove(FoodFile)

		if err := LogInteractive(); err == nil {
			t.Fatal("expected an error")
		}
		if exists(path) {
			t.Error("nothing should be saved")
		}
	})

	t.Run("empty food list is an error", func(t *testing.T) {
		path := useTempData(t)
		writeJSON(FoodFile, []Food{})

		if err := LogInteractive(); err == nil {
			t.Fatal("expected an error")
		}
		if exists(path) {
			t.Error("nothing should be saved")
		}
	})

	t.Run("broken food file is an error", func(t *testing.T) {
		path := useTempData(t)
		os.WriteFile(FoodFile, []byte("not json"), 0o644)

		if err := LogInteractive(); err == nil {
			t.Fatal("expected an error")
		}
		if got := readFile(t, FoodFile); got != "not json" {
			t.Errorf("food file was overwritten with: %s", got)
		}
		if exists(path) {
			t.Error("nothing should be saved")
		}
	})
}

// ============================================================
// UndoLog
// ============================================================

func TestUndoLog(t *testing.T) {
	t.Run("removes only the last entry", func(t *testing.T) {
		path := useTempData(t)
		seedLogs(t)

		if err := UndoLog(); err != nil {
			t.Fatal(err)
		}

		logs, err := LoadLog(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(logs) != 2 || logs[0].Food != "Oats" || logs[1].Food != "Cheese" {
			t.Errorf("want [Oats Cheese], got %+v", logs)
		}
	})

	t.Run("twice removes two entries", func(t *testing.T) {
		path := useTempData(t)
		seedLogs(t)

		UndoLog()
		if err := UndoLog(); err != nil {
			t.Fatal(err)
		}

		logs, _ := LoadLog(path)
		if len(logs) != 1 || logs[0].Food != "Oats" {
			t.Errorf("want [Oats], got %+v", logs)
		}
	})

	t.Run("last entry leaves an empty list", func(t *testing.T) {
		path := useTempData(t)
		writeJSON(LogFile, []Log{{Time: time.Now(), Food: "Oats", Grams: 50}})

		if err := UndoLog(); err != nil {
			t.Fatal(err)
		}

		logs, err := LoadLog(path)
		if err != nil {
			t.Fatalf("the file should still load: %v", err)
		}
		if len(logs) != 0 {
			t.Errorf("got %d entries, want 0", len(logs))
		}

		if err := UndoLog(); err == nil {
			t.Error("a second undo on the empty log should give an error")
		}
	})

	t.Run("nothing to undo is an error", func(t *testing.T) {
		path := useTempData(t) // no log file yet

		if err := UndoLog(); err == nil {
			t.Fatal("expected an error")
		}
		if exists(path) {
			t.Error("undo should not create a log file")
		}
	})

	t.Run("broken file is not overwritten", func(t *testing.T) {
		path := useTempData(t)
		os.WriteFile(path, []byte("not json"), 0o644)

		if err := UndoLog(); err == nil {
			t.Fatal("expected an error")
		}
		if got := readFile(t, path); got != "not json" {
			t.Errorf("broken file was overwritten with: %s", got)
		}
	})

	t.Run("does not touch the food file", func(t *testing.T) {
		useTempData(t)
		seedLogs(t)
		before := readFile(t, FoodFile)

		if err := UndoLog(); err != nil {
			t.Fatal(err)
		}
		if got := readFile(t, FoodFile); got != before {
			t.Errorf("food file changed:\nbefore: %s\nafter: %s", before, got)
		}
	})
}

// ============================================================
// UndoFood
// ============================================================

func TestUndoFood(t *testing.T) {
	t.Run("removes only the last food", func(t *testing.T) {
		useTempData(t) // saves Oats, Cheese, Peanut butter

		if err := UndoFood(); err != nil {
			t.Fatal(err)
		}

		foods, err := LoadFood(FoodFile)
		if err != nil {
			t.Fatal(err)
		}
		if len(foods) != 2 || foods[0].Name != "Oats" || foods[1].Name != "Cheese" {
			t.Errorf("want [Oats Cheese], got %+v", foods)
		}
	})

	t.Run("nothing to undo is an error", func(t *testing.T) {
		useTempData(t)
		os.Remove(FoodFile)

		if err := UndoFood(); err == nil {
			t.Fatal("expected an error")
		}
		if exists(FoodFile) {
			t.Error("undo should not create a food file")
		}
	})

	t.Run("broken file is not overwritten", func(t *testing.T) {
		useTempData(t)
		os.WriteFile(FoodFile, []byte("not json"), 0o644)

		if err := UndoFood(); err == nil {
			t.Fatal("expected an error")
		}
		if got := readFile(t, FoodFile); got != "not json" {
			t.Errorf("broken file was overwritten with: %s", got)
		}
	})

	t.Run("does not touch the log file", func(t *testing.T) {
		useTempData(t)
		seedLogs(t)
		before := readFile(t, LogFile)

		if err := UndoFood(); err != nil {
			t.Fatal(err)
		}
		if got := readFile(t, LogFile); got != before {
			t.Errorf("log file changed:\nbefore: %s\nafter: %s", before, got)
		}

		// and the food file really did change, so this is not passing by accident
		foods, _ := LoadFood(FoodFile)
		if len(foods) != 2 {
			t.Errorf("want 2 foods left, got %d", len(foods))
		}
	})
}

// ============================================================
// Helpers (used by the tests above)
// ============================================================

// useTempData points FoodFile and LogFile at a temp folder, puts a few foods in
// the food file, and fakes the USB drive (user answers No to "Mount?"),
// so tests never touch your real data. It returns the log file path.
func useTempData(t *testing.T) string {
	t.Helper()

	oldFood, oldFoodBackup := FoodFile, FoodFileBackup
	oldLog, oldLogBackup := LogFile, LogFileBackup
	t.Cleanup(func() {
		FoodFile, FoodFileBackup = oldFood, oldFoodBackup
		LogFile, LogFileBackup = oldLog, oldLogBackup
	})

	dir := t.TempDir()
	FoodFile = filepath.Join(dir, "local", "food.json")
	FoodFileBackup = filepath.Join(dir, "usb", "food.json")
	LogFile = filepath.Join(dir, "local", "log.json")
	LogFileBackup = filepath.Join(dir, "usb", "log.json")
	fakeDrive(t, false, nil, false)
	quiet(t)

	foods := []Food{{Name: "Oats", Kcal: 370}, {Name: "Cheese", Kcal: 400}, {Name: "Peanut butter", Kcal: 600}}
	if err := writeJSON(FoodFile, foods); err != nil {
		t.Fatal(err)
	}

	return LogFile
}

// withArgs sets os.Args like running: go run . <args...>
func withArgs(t *testing.T, args ...string) {
	t.Helper()

	old := os.Args
	os.Args = append([]string{"food"}, args...)
	t.Cleanup(func() { os.Args = old })
}

// seedLogs writes three log entries into LogFile (call after useTempData).
func seedLogs(t *testing.T) {
	t.Helper()

	now := time.Now()
	logs := []Log{
		{Time: now, Food: "Oats", Grams: 50},
		{Time: now, Food: "Cheese", Grams: 30},
		{Time: now, Food: "Peanut butter", Grams: 20},
	}
	if err := writeJSON(LogFile, logs); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// fakeDrive sets how the USB drive behaves and restores the real functions after the test.
func fakeDrive(t *testing.T, mounted bool, checkErr error, mountAnswer bool) {
	t.Helper()

	oldCheck, oldAsk := checkMounted, askMount
	t.Cleanup(func() { checkMounted, askMount = oldCheck, oldAsk })

	checkMounted = func() (bool, error) { return mounted, checkErr }
	askMount = func() (bool, error) { return mountAnswer, nil }
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// quiet hides what the app itself prints (prompts, "Local save!", ...) while
// the test runs, so the test output stays readable. Failures still show.
func quiet(t *testing.T) {
	t.Helper()

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = devNull
	t.Cleanup(func() {
		os.Stdout = old
		devNull.Close()
	})
}
