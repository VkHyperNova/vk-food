package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Food holds nutrition values per 100 g.
type Food struct {
	Name    string  `json:"name"`
	Kcal    float64 `json:"kcal"`
	Protein float64 `json:"protein"`
	Carbs   float64 `json:"carbs"`
	Fat     float64 `json:"fat"`
	Fiber   float64 `json:"fiber"`
	Sugar   float64 `json:"sugar"`
	Salt    float64 `json:"salt"`
}

// Entry is one thing you ate: a food name and an amount in grams.
type Entry struct {
	Time  time.Time `json:"time"`
	Food  string    `json:"food"`
	Grams float64   `json:"grams"`
}

// Data is everything saved to disk.
type Data struct {
	Foods   []Food  `json:"foods"`
	Entries []Entry `json:"entries"`
}

// Totals is a sum of nutrients (absolute amounts, not per 100 g).
type Totals struct {
	Kcal, Protein, Carbs, Fat, Fiber, Sugar, Salt float64
}

// Add adds the nutrients of grams of food f to the totals.
func (t *Totals) Add(f Food, grams float64) {
	k := grams / 100
	t.Kcal += f.Kcal * k
	t.Protein += f.Protein * k
	t.Carbs += f.Carbs * k
	t.Fat += f.Fat * k
	t.Fiber += f.Fiber * k
	t.Sugar += f.Sugar * k
	t.Salt += f.Salt * k
}

// dataPath returns where the JSON file lives.
// Override with the FOODTRACKER_DATA environment variable.
func dataPath() (string, error) {
	if p := os.Getenv("FOODTRACKER_DATA"); p != "" {
		return p, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "foodtracker", "data.json"), nil
}

func load() (*Data, error) {
	path, err := dataPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Data{}, nil // first run
	}
	if err != nil {
		return nil, err
	}
	var d Data
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return &d, nil
}

func (d *Data) save() error {
	path, err := dataPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	// Write to a temp file first so a crash can't corrupt your data.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// findFood looks up a food by name: exact match first (case-insensitive),
// then a unique partial match.
func (d *Data) findFood(query string) (Food, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	var partial []Food
	for _, f := range d.Foods {
		name := strings.ToLower(f.Name)
		if name == q {
			return f, nil
		}
		if strings.Contains(name, q) {
			partial = append(partial, f)
		}
	}
	switch len(partial) {
	case 0:
		return Food{}, fmt.Errorf("no food named %q (add it with: food add-food %q --kcal ...)", query, query)
	case 1:
		return partial[0], nil
	}
	names := make([]string, len(partial))
	for i, f := range partial {
		names[i] = f.Name
	}
	return Food{}, fmt.Errorf("%q matches several foods: %s", query, strings.Join(names, ", "))
}

// upsertFood adds a food or replaces one with the same name.
func (d *Data) upsertFood(f Food) (replaced bool) {
	for i, existing := range d.Foods {
		if strings.EqualFold(existing.Name, f.Name) {
			d.Foods[i] = f
			return true
		}
	}
	d.Foods = append(d.Foods, f)
	return false
}
