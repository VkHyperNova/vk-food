package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

const (
	driveLabel      = "VK DATA"
	driveMountPoint = "/media/veikko/VK DATA"
	BaseDB          = "FOOD"
	BaseLocal       = "DATABASES"
)

// First file
var FoodFile = filepath.Join(BaseLocal, BaseDB, "food.json")
var FoodFileBackup = filepath.Join(driveMountPoint, BaseLocal, BaseDB, "food.json")

// Second file
var LogFile = filepath.Join(BaseLocal, BaseDB, "log.json")
var LogFileBackup = filepath.Join(driveMountPoint, BaseLocal, BaseDB, "log.json")

var ErrBackupSkipped = errors.New("backup skipped")

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
	if len(os.Args) < 2 {
		Report(foods, logs)
		return
	}

	switch os.Args[1] {
	case "add-food":
		if err := AddFood(foods); err != nil {
			fmt.Println("error:", err)
		}
	default:
		fmt.Println("unknown command:", os.Args[1])
	}
}
