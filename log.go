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
	return Save(LogFile, LogFileBackup, logs)
}

func AddLog() error {

	if len(os.Args) < 4 {
		return errors.New("usage: go run . log <food-name> <grams>")
	}

	grams, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		return fmt.Errorf("grams must be a valid number: %w", err)
	}

	if grams <= 0 {
		return errors.New("grams must be greater than zero")
	}

	foods, err := LoadFood(FoodFile)
	if err != nil {
		return fmt.Errorf("load food: %w", err)
	}

	food, ok := findFood(foods, os.Args[2])
	if !ok {
		return fmt.Errorf("%q is not in the food list, add it first with: go run . add-food", os.Args[2])
	}

	logs, err := LoadLog(LogFile)
	if err != nil {
		return fmt.Errorf("load log: %w", err)
	}

	logs = append(logs, Log{Time: time.Now(), Food: food.Name, Grams: grams})

	return SaveLog(logs)
}

func UndoLog() error {
	logs, err := LoadLog(LogFile)
	if err != nil {
		return fmt.Errorf("load log: %w", err)
	}
	if len(logs) == 0 {
		return errors.New("nothing to undo")
	}

	return SaveLog(logs[:len(logs)-1])
}
