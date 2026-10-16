package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// useTempLogFile points LogFile at a temp folder and fakes the USB drive
// (user answers No to "Mount?"), so tests never touch your real data.
func useTempLogFile(t *testing.T) string {
	t.Helper()

	oldFile, oldBackup := LogFile, LogFileBackup
	t.Cleanup(func() { LogFile, LogFileBackup = oldFile, oldBackup })

	dir := t.TempDir()
	LogFile = filepath.Join(dir, "local", "log.json")
	LogFileBackup = filepath.Join(dir, "usb", "log.json")
	fakeDrive(t, false, nil, false)

	return LogFile
}

// withArgs sets os.Args like running: go run . <args...>
func withArgs(t *testing.T, args ...string) {
	t.Helper()

	old := os.Args
	os.Args = append([]string{"food"}, args...)
	t.Cleanup(func() { os.Args = old })
}

func TestLoadLogMissingFile(t *testing.T) {
	logs, err := LoadLog(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatalf("missing file should not be an error, got %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("got %d entries, want 0", len(logs))
	}
}

func TestLoadLogBadJSON(t *testing.T) {
	for _, content := range []string{"not json", `{}`} {
		path := filepath.Join(t.TempDir(), "log.json")
		os.WriteFile(path, []byte(content), 0o644)

		if _, err := LoadLog(path); err == nil {
			t.Errorf("content %q should give an error", content)
		}
	}
}

func TestSaveLogRoundtrip(t *testing.T) {
	path := useTempLogFile(t)
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
}

func TestAddLogFirstRun(t *testing.T) {
	path := useTempLogFile(t)
	withArgs(t, "log", "Oats", "50")

	before := time.Now()
	if err := AddLog(); err != nil {
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
}

func TestAddLogAppends(t *testing.T) {
	path := useTempLogFile(t)

	withArgs(t, "log", "Oats", "50")
	if err := AddLog(); err != nil {
		t.Fatal(err)
	}
	withArgs(t, "log", "Cheese", "30")
	if err := AddLog(); err != nil {
		t.Fatal(err)
	}

	logs, err := LoadLog(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 || logs[0].Food != "Oats" || logs[1].Food != "Cheese" {
		t.Errorf("old entry lost or order wrong: %+v", logs)
	}
}

func TestAddLogNameWithSpaces(t *testing.T) {
	path := useTempLogFile(t)
	withArgs(t, "log", "peanut butter", "30") // the shell passes quoted text as one argument

	if err := AddLog(); err != nil {
		t.Fatal(err)
	}

	logs, _ := LoadLog(path)
	if len(logs) != 1 || logs[0].Food != "peanut butter" {
		t.Errorf("unexpected logs: %+v", logs)
	}
}

func TestAddLogBadInput(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no arguments", []string{"log"}},
		{"missing grams", []string{"log", "Oats"}},
		{"grams not a number", []string{"log", "Oats", "abc"}},
		{"grams with comma", []string{"log", "Oats", "1,5"}},
		{"zero grams", []string{"log", "Oats", "0"}},
		{"negative grams", []string{"log", "Oats", "-5"}},
		{"NaN grams", []string{"log", "Oats", "NaN"}},
		{"infinite grams", []string{"log", "Oats", "Inf"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := useTempLogFile(t)
			withArgs(t, tc.args...)

			if err := AddLog(); err == nil {
				t.Error("expected an error")
			}
			if exists(path) {
				t.Error("nothing should be saved after bad input")
			}
		})
	}
}

func TestAddLogBrokenFileIsNotOverwritten(t *testing.T) {
	path := useTempLogFile(t)
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte("not json"), 0o644)
	withArgs(t, "log", "Oats", "50")

	if err := AddLog(); err == nil {
		t.Fatal("expected an error for the broken file")
	}

	data, _ := os.ReadFile(path)
	if string(data) != "not json" {
		t.Errorf("broken file was overwritten with: %s", data)
	}
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
