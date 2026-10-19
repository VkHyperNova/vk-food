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

/* Commands*/

func AddFood() error {

	foods, err := LoadFood(FoodFile)
	if err != nil {
		return fmt.Errorf("load food: %w", err)
	}

	newFood := askFood()
	if err := checkName(foods, newFood.Name, -1); err != nil {
		return err
	}

	foods = append(foods, newFood)

	return SaveFood(foods)
}

func UndoFood() error {
	foods, err := LoadFood(FoodFile)
	if err != nil {
		return fmt.Errorf("load log: %w", err)
	}
	if len(foods) == 0 {
		return errors.New("nothing to undo")
	}

	return SaveFood(foods[:len(foods)-1])
}

func DeleteFood() error {
	if len(os.Args) < 3 {
		return errors.New("usage: go run . delete-food <food-name>")
	}
	foods, err := LoadFood(FoodFile)
	if err != nil {
		return fmt.Errorf("load food: %w", err)
	}
	i := indexOfFood(foods, os.Args[2])
	if i < 0 {
		return fmt.Errorf("%q is not in the food list", os.Args[2])
	}
	logs, err := LoadLog(LogFile)
	if err != nil {
		return fmt.Errorf("load log: %w", err)
	}
	for _, l := range logs {
		if strings.EqualFold(l.Food, foods[i].Name) {
			return fmt.Errorf("%q is used in the log, undo those entries first", foods[i].Name)
		}
	}
	return SaveFood(append(foods[:i], foods[i+1:]...))
}

func UpdateFood() error {
	// os.Args is: [program, "update-food", "<food-name>"], so we need 3 items
	if len(os.Args) < 3 {
		return errors.New("usage: go run . update-food <food-name>")
	}

	// read food.json into a slice
	foods, err := LoadFood(FoodFile)
	if err != nil {
		return fmt.Errorf("load food: %w", err)
	}

	// find the position of the food the user typed (ignores upper/lower case)
	i := indexOfFood(foods, os.Args[2])
	if i < 0 { // -1 means not found
		return fmt.Errorf("%q is not in the food list", os.Args[2])
	}

	// read log.json too, because a rename must also change old log entries
	logs, err := LoadLog(LogFile)
	if err != nil {
		return fmt.Errorf("load log: %w", err)
	}

	// ask for new values, showing the current ones (Enter keeps them)
	updated := askFoodKeep(foods[i])

	// name must not be empty and must not belong to a different food
	// (i is passed so the food can keep its own name)
	if err := checkName(foods, updated.Name, i); err != nil {
		return err
	}

	// remember the old name, we need it to find the matching log entries
	oldName := foods[i].Name

	// replace the old food with the updated one
	foods[i] = updated

	// go through every log entry
	for k := range logs {
		// if the entry uses the old name, switch it to the new name
		if strings.EqualFold(logs[k].Food, oldName) {
			logs[k].Food = updated.Name
		}
	}

	// write food.json (and the backup)
	if err := SaveFood(foods); err != nil {
		return err
	}

	// write log.json (and the backup)
	return SaveLog(logs)
}

/* Storage */

func SaveFood(foods []Food) error {
	return Save(FoodFile, FoodFileBackup, foods)
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

/* Helper functions */

func askFood() Food {
	var f Food

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Food name: ")
	f.Name, _ = reader.ReadString('\n')
	f.Name = strings.TrimSpace(f.Name)

	f.Kcal = askFloat(reader, "Calories (kcal): ")
	f.Protein = askFloat(reader, "Protein (g): ")
	f.Carbs = askFloat(reader, "Carbs (g): ")
	f.Fat = askFloat(reader, "Fat (g): ")
	f.Fiber = askFloat(reader, "Fiber (g): ")
	f.Sugar = askFloat(reader, "Sugar (g): ")
	f.Salt = askFloat(reader, "Salt (g): ")

	return f
}

func askFoodKeep(current Food) Food {
	reader := bufio.NewReader(os.Stdin)
	f := current

	fmt.Printf("Food name [%s%s%s]: ", Green, current.Name, Reset)
	name, _ := reader.ReadString('\n')
	if name = strings.TrimSpace(name); name != "" {
		f.Name = name
	}

	f.Kcal = askFloatKeep(reader, "Calories (kcal)", current.Kcal)
	f.Protein = askFloatKeep(reader, "Protein (g)", current.Protein)
	f.Carbs = askFloatKeep(reader, "Carbs (g)", current.Carbs)
	f.Fat = askFloatKeep(reader, "Fat (g)", current.Fat)
	f.Fiber = askFloatKeep(reader, "Fiber (g)", current.Fiber)
	f.Sugar = askFloatKeep(reader, "Sugar (g)", current.Sugar)
	f.Salt = askFloatKeep(reader, "Salt (g)", current.Salt)

	return f
}

func askFloatKeep(reader *bufio.Reader, label string, current float64) float64 {
	return readFloat(reader, fmt.Sprintf("%s [%s%g%s]: ", label, Green, current, Reset), &current)
}

func readFloat(reader *bufio.Reader, prompt string, keep *float64) float64 {
	for {
		fmt.Print(prompt)

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Input error:", err)
			os.Exit(1)
		}

		input = strings.TrimSpace(input)

		if input == "" && keep != nil {
			return *keep
		}

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

func indexOfFood(foods []Food, name string) int {
	for i, f := range foods {
		if strings.EqualFold(f.Name, name) {
			return i
		}
	}
	return -1
}

func checkName(foods []Food, name string, skip int) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if i := indexOfFood(foods, name); i >= 0 && i != skip {
		return fmt.Errorf("%q already exists", name)
	}
	return nil
}

func askFloat(reader *bufio.Reader, prompt string) float64 {
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
