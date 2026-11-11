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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

type Log struct {
	Time  time.Time `json:"time"`
	Food  string    `json:"food"`
	Grams float64   `json:"grams"`
}

/* Commands */

func AddLog() error {
	if len(os.Args) < 3 {
		return logInteractive()
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

/* Helper Functions */

func logInteractive() error {
	foods, err := LoadFood(Foods)
	if err != nil {
		return fmt.Errorf("load food: %w", err)
	}
	if len(foods) == 0 {
		return errors.New("no foods yet, add one first with: go run . add-food")
	}

	// Terminal width, so no line is ever wider than the window
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		w = 80
	}

	options := foodOptions(foods, w)

	var name, gramsText string
	form := huh.NewForm(
		// Page 1: pick the food (type to filter)
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Food").
				Options(options...).
				Filtering(true). // list narrows as you type
				Height(10).      // fixed height, scrolls instead of filling the screen
				Value(&name),
		),
		// Page 2: grams, shown after the food is chosen
		huh.NewGroup(
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
		),
	).
		WithWidth(w).
		WithProgramOptions(tea.WithAltScreen()). // alt screen so a stale frame can't be left behind
		WithShowHelp(w >= 75)                    // hide the hint bar when it would wrap

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil // Esc or Ctrl-C is not an error
		}
		return err
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
