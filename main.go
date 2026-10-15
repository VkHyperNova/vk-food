package main

import (
	"fmt"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	Bold   = "\033[1m"
	Italic = "\033[3m"
)

func main() {

	foods, err := LoadFood(FoodFile)
	if err != nil {
		fmt.Printf("LoadFood: %v\n", err)
	}

	logs, err := LoadLog(LogFile)
	if err != nil {
		fmt.Printf("LoadLog: %v\n", err)
	}

	RunApp(foods, logs)
}

func RunApp(foods []Food, logs []Log) {
	for {
		RunStats(foods, logs)

		var command string

		fmt.Scanln(&command)

		switch command {
		case "add-food":
			AskFood()
		case "log":
		case "undo":

		}

	}

}
