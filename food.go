package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Food holds nutrition values per 100 g.
type Food struct {
	Name    string
	Kcal    float64
	Protein float64
	Carbs   float64
	Fat     float64
	Fiber   float64
	Sugar   float64
	Salt    float64
}

// writeJSON saves any value to a file as indented JSON.
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func writeFoods(path string, foods []Food) error {
	return writeJSON(path, foods)
}

func SaveToFile(foods []Food) error {
	// Save local
	if err := writeFoods(FoodFile, foods); err != nil {
		return err
	}
	fmt.Println(Green + "Local save!" + Reset)

	// Save backup (best effort: warn, but don't fail the whole save)
	if err := InitBackupStorage(); err != nil {
		fmt.Println(Yellow + "Backup init failed: " + err.Error() + Reset)
		return nil
	}
	if err := writeFoods(FoodFileBackup, foods); err != nil {
		fmt.Println(Yellow + "Backup write failed: " + err.Error() + Reset)
		return nil
	}
	fmt.Println(Green + "Backup save!" + Reset)

	return nil
}

// LoadFood reads the list of foods from a JSON file.
func LoadFood(source string) ([]Food, error) {
	data, err := os.ReadFile(source)
	if err != nil {
		return nil, err
	}
	var foods []Food
	if err := json.Unmarshal(data, &foods); err != nil {
		return nil, err
	}
	return foods, nil
}
