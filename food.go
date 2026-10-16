package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
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

func AskFloat(reader *bufio.Reader, prompt string) float64 {
	for {
		fmt.Print(prompt)

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Input error:", err)
			os.Exit(1)
		}

		input = strings.TrimSpace(input)

		value, err := strconv.ParseFloat(input, 64)
		if err != nil {
			fmt.Println("Invalid number. Please try again.")
			continue
		}

		if value < 0 || math.IsInf(value, 0) || math.IsNaN(value) {
			fmt.Println("Please enter a valid non-negative number.")
			continue
		}

		return value
	}
}

func AskFood() Food {
	var f Food

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Food name: ")
	f.Name, _ = reader.ReadString('\n')
	f.Name = strings.TrimSpace(f.Name)

	f.Kcal = AskFloat(reader, "Calories (kcal): ")
	f.Protein = AskFloat(reader, "Protein (g): ")
	f.Carbs = AskFloat(reader, "Carbs (g): ")
	f.Fat = AskFloat(reader, "Fat (g): ")
	f.Fiber = AskFloat(reader, "Fiber (g): ")
	f.Sugar = AskFloat(reader, "Sugar (g): ")
	f.Salt = AskFloat(reader, "Salt (g): ")

	return f
}

func AddFood() error {

	foods, err := LoadFood(FoodFile)
	if err != nil {
		return fmt.Errorf("load food: %w", err)
	}

	newFood := AskFood()

	foods = append(foods, newFood)

	return SaveFood(foods)
}

func SaveFood(foods []Food) error {
	return SaveWithBackup(FoodFile, FoodFileBackup, foods)
}

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
