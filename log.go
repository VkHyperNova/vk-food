package main

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

// Log is one thing you ate: which food, how much, and when.
type Log struct {
	Time  time.Time `json:"time"`
	Food  string    `json:"food"`
	Grams float64   `json:"grams"`
}

// LoadLog reads all log entries from a JSON file if it exists
func LoadLog(source string) ([]Log, error) {
	data, err := os.ReadFile(source)

	// If file does not exist continue
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil // first run: nothing saved yet
	}

	if err != nil {
		return nil, err
	}
	var entries []Log
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func SaveLog(logs []Log) error {
	return saveWithBackup(LogFile, LogFileBackup, logs)
}

// AddLog returns the log with one more entry, timestamped now.
func AddLog(logs []Log, food string, grams float64) []Log {
	return append(logs, Log{Time: time.Now(), Food: food, Grams: grams})
}

// sameDay reports whether two times fall on the same calendar day (local time).
func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Local().Date()
	by, bm, bd := b.Local().Date()
	return ay == by && am == bm && ad == bd
}

// LogsOn returns the entries logged on the given day.
func LogsOn(entries []Log, day time.Time) []Log {
	var out []Log
	for _, e := range entries {
		if sameDay(e.Time, day) {
			out = append(out, e)
		}
	}
	return out
}

// TodayLogs returns only the entries logged today.
func TodayLogs(entries []Log) []Log {
	return LogsOn(entries, time.Now())
}
