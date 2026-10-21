package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
)

type Log struct {
	Time  time.Time `json:"time"`
	Food  string    `json:"food"`
	Grams float64   `json:"grams"`
}

/* Commands */

func AddLog() error {
	if len(os.Args) < 3 {
		return LogInteractive()
	}
	if len(os.Args) < 4 {
		return errors.New("usage: go run . log <food-name> <grams>")
	}
	grams, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		return fmt.Errorf("grams must be a valid number: %w", err)
	}
	return logFood(os.Args[2], grams)
}

func UndoLog() error {
	logs, err := LoadLog(Logs)
	if err != nil {
		return fmt.Errorf("load log: %w", err)
	}
	if len(logs) == 0 {
		return errors.New("nothing to undo")
	}

	return SaveLog(logs[:len(logs)-1])
}

func LogInteractive() error {
	foods, err := LoadFood(Foods)
	if err != nil {
		return fmt.Errorf("load food: %w", err)
	}
	if len(foods) == 0 {
		return errors.New("no foods yet, add one first with: go run . add-food")
	}

	options := make([]huh.Option[string], len(foods))
	for i, f := range foods {
		label := fmt.Sprintf("%s (%v kcal)", f.Name, f.Kcal)
		options[i] = huh.NewOption(label, f.Name)
	}

	var name, gramsText string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Food").
			Options(options...).
			Filtering(true). // typing narrows the list
			Value(&name),
		huh.NewInput().
			Title("Grams").
			Value(&gramsText).
			Validate(func(s string) error {
				g, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
				if err != nil || g <= 0 {
					return errors.New("enter a number greater than zero")
				}
				return nil
			}),
	))
	if err := form.Run(); err != nil {
		return err // pressing Esc or Ctrl-C gives huh.ErrUserAborted
	}

	grams, _ := strconv.ParseFloat(strings.TrimSpace(gramsText), 64)
	return logFood(name, grams)
}

func logFood(name string, grams float64) error {
	if grams <= 0 || math.IsNaN(grams) || math.IsInf(grams, 0) {
		return errors.New("grams must be a valid number greater than zero")
	}
	foods, err := LoadFood(Foods)
	if err != nil {
		return fmt.Errorf("load food: %w", err)
	}
	food, ok := findFood(foods, name)
	if !ok {
		return fmt.Errorf("%q is not in the food list, add it first with: go run . add-food", name)
	}
	logs, err := LoadLog(Logs)
	if err != nil {
		return fmt.Errorf("load log: %w", err)
	}
	logs = append(logs, Log{Time: time.Now(), Food: food.Name, Grams: grams})
	return SaveLog(logs)
}

/* Storage */

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
	return Save(Logs, LogsBackup, logs)
}
