package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Entry is one thing you ate: which food, how much, and when.
type Entry struct {
	Time  time.Time `json:"time"`
	Food  string    `json:"food"`
	Grams float64   `json:"grams"`
}

func writeLogs(path string, entries []Entry) error {
	return writeJSON(path, entries)
}

// LoadLog reads all log entries from a JSON file.
func LoadLog(source string) ([]Entry, error) {
	data, err := os.ReadFile(source)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// SaveLog writes all log entries to the local log file.
func SaveLog(entries []Entry) error {
	if err := writeLogs(LogFile, entries); err != nil {
		return err
	}
	fmt.Println(Green + "Log saved!" + Reset)
	return nil
}

// AddEntry returns the log with one more entry, timestamped now.
func AddEntry(entries []Entry, food string, grams float64) []Entry {
	return append(entries, Entry{Time: time.Now(), Food: food, Grams: grams})
}

// sameDay reports whether two times fall on the same calendar day (local time).
func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Local().Date()
	by, bm, bd := b.Local().Date()
	return ay == by && am == bm && ad == bd
}

// EntriesOn returns the entries logged on the given day.
func EntriesOn(entries []Entry, day time.Time) []Entry {
	var out []Entry
	for _, e := range entries {
		if sameDay(e.Time, day) {
			out = append(out, e)
		}
	}
	return out
}

// TodayEntries returns only the entries logged today.
func TodayEntries(entries []Entry) []Entry {
	return EntriesOn(entries, time.Now())
}
