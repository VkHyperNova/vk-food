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
	// BaseLocal = "DEVELOPMENT"
)

var FoodFile = filepath.Join(BaseLocal, BaseDB, "foods.json")
var LogFile = filepath.Join(BaseLocal, BaseDB, "logs.json")

// Backup on D drive \\
var FoodFileBackup = filepath.Join(driveMountPoint, BaseLocal, BaseDB, "foods.json")
var LogFileBackup = filepath.Join(driveMountPoint, BaseLocal, BaseDB, "logs.json")

var ErrBackupSkipped = errors.New("backup skipped")

const usage = `Usage: go run . <command>

Commands:
  (none)                      show the report
  log                         pick a food and enter grams
  log <food> <grams>          log a food, e.g. log oats 50
  undo                        remove the last log entry
  add-food                    add a new food (values per 100 g)
  update-food <food>          change a food (Enter keeps the old value)
  delete-food <food>          delete a food that is not in the log
  undo-food                   remove the last added food
  help                        show this text`

func printUsage() {
	fmt.Println(usage)
}

func main() {
	var err error

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			printUsage()
			return
		case "add-food":
			err = AddFood()
		case "undofood":
			err = UndoFood()
		case "update-food":
			err = UpdateFood()
		case "delete-food":
			err = DeleteFood()
		case "log":
			err = AddLog()
		case "undolog":
			err = UndoLog()
		default:
			err = fmt.Errorf("unknown command: %s", os.Args[1])
		}
	}

	Report()

	if err != nil {
		fmt.Printf("%serror: %v%s\n", Red, err, Reset)
	}
}
