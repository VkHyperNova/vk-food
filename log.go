package main

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

type Log struct {
	Time  time.Time `json:"time"`
	Food  string    `json:"food"`
	Grams float64   `json:"grams"`
}

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
	return SaveWithBackup(LogFile, LogFileBackup, logs)
}

func AddLog(logs []Log, food string, grams float64) error {
	logs = append(logs, Log{Time: time.Now(), Food: food, Grams: grams})
	return SaveLog(logs)
}
