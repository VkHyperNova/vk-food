package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
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

func AddLog() error {

	if len(os.Args) < 4 {
		fmt.Println("Usage: go run . log <food-name> <grams>")
		return errors.New("usage: go run . log <food-name> <grams>")
	}

	grams, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		fmt.Println("error: grams must be a valid number")
		return fmt.Errorf("grams must be a valid number: %w", err)
	}

	if grams <= 0 {
		fmt.Println("error: grams must be greater than zero")
		return errors.New("grams must be greater than zero")
	}

	logs, err := LoadLog(LogFile)
	if err != nil {
		return fmt.Errorf("load log: %w", err)
	}

	logs = append(logs, Log{Time: time.Now(), Food: os.Args[2], Grams: grams})

	return SaveLog(logs)
}
