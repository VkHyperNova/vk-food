package main

import (
	"fmt"
	"log"
	"time"
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

	defer CloseInput()

	if err := InitLocalStorage(); err != nil {
		log.Fatalf("init storage: %v", err)
	}

	// Load foods
	foods, err := LoadFood(FoodFile)
	if err != nil {
		fmt.Printf("LoadFood: %v\n", err)
	}

	entries, err := LoadLog(LogFile)
	if err != nil {
		fmt.Printf("LoadLog: %v\n", err)
	}

	Run(foods, entries)

}

func Run(foods []Food, entries []Entry) {
	today := time.Now()
	printWeek(foods, entries, today)
	fmt.Println()
	printToday(foods, entries, today)
}
