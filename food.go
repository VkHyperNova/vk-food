package main

import (
	"encoding/json"
	"errors"
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

func AskFood() Food {
	var f Food

	fmt.Print("Food name: ")
	fmt.Scanln(&f.Name)

	fmt.Print("Calories (kcal): ")
	fmt.Scanln(&f.Kcal)

	fmt.Print("Protein (g): ")
	fmt.Scanln(&f.Protein)

	fmt.Print("Carbs (g): ")
	fmt.Scanln(&f.Carbs)

	fmt.Print("Fat (g): ")
	fmt.Scanln(&f.Fat)

	fmt.Print("Fiber (g): ")
	fmt.Scanln(&f.Fiber)

	fmt.Print("Sugar (g): ")
	fmt.Scanln(&f.Sugar)

	fmt.Print("Salt (g): ")
	fmt.Scanln(&f.Salt)

	return f
}

func SaveFood(foods []Food) error {
	return saveWithBackup(FoodFile, FoodFileBackup, foods)
}

// LoadFood reads the list of foods from a JSON file if it exists
func LoadFood(source string) ([]Food, error) {
	data, err := os.ReadFile(source)

	// If file does not exist continue
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil // first run: nothing saved yet
	}
	if err != nil {
		return nil, err
	}
	var foods []Food
	if err := json.Unmarshal(data, &foods); err != nil {
		return nil, err
	}
	return foods, nil
}
