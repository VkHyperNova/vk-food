package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAddEntry(t *testing.T) {
	var entries []Entry
	entries = AddEntry(entries, "Oats", 80)
	entries = AddEntry(entries, "Banana", 120)

	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].Food != "Oats" || entries[0].Grams != 80 {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
	if entries[0].Time.IsZero() {
		t.Error("entry should have a timestamp")
	}
}

func TestTodayEntries(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		{Time: now.AddDate(0, 0, -1), Food: "Yesterday", Grams: 100},
		{Time: now, Food: "Today", Grams: 100},
		{Time: now.AddDate(0, 0, 1), Food: "Tomorrow", Grams: 100},
	}

	got := TodayEntries(entries)
	if len(got) != 1 || got[0].Food != "Today" {
		t.Errorf("expected only today's entry, got %+v", got)
	}
}

func TestWriteAndLoadLog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log.json")
	want := AddEntry(nil, "Oats", 80)

	if err := writeLogs(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadLog(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Food != "Oats" || got[0].Grams != 80 {
		t.Fatalf("got %+v", got)
	}
	if !got[0].Time.Equal(want[0].Time) {
		t.Errorf("time changed: got %v, want %v", got[0].Time, want[0].Time)
	}
}

func TestLoadEmptyLog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log.json")
	if err := os.WriteFile(path, []byte("[]"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadLog(path)
	if err != nil {
		t.Fatalf("an empty log should load, got: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no entries, got %d", len(got))
	}
}

func TestLoadLogMissingFile(t *testing.T) {
	_, err := LoadLog(filepath.Join(t.TempDir(), "nope.json"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected ErrNotExist, got %v", err)
	}
}
