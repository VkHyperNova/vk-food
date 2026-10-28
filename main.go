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
	FoodsFile       = "foods.json"
	LogsFile        = "logs.json"
	BaseLocal       = "DATABASES"
	// BaseLocal = "DEVELOPMENT"
)

var Foods = filepath.Join(BaseLocal, BaseDB, FoodsFile)
var Logs = filepath.Join(BaseLocal, BaseDB, LogsFile)

// Backup on D drive \\
var FoodsBackup = filepath.Join(driveMountPoint, BaseLocal, BaseDB, FoodsFile)
var LogsBackup = filepath.Join(driveMountPoint, BaseLocal, BaseDB, LogsFile)

var ErrBackupSkipped = errors.New("backup skipped")

const usage = `Usage: go run . <command>

Commands:
  (none)                      show the report
  log                         pick a food and enter grams
  log <food> <grams>          log a food, e.g. log oats 50
  undo-log                    remove the last log entry
  add-food                    add a new food (values per 100 g)
  update-food <food>          change a food (Enter keeps the old value)
  delete-food <food>          delete a food that is not in the log
  undo-food                   remove the last added food
  help                        show this text
  print-foods                 show all foods with their nutrition values`

func main() {
	var err error

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			fmt.Println(usage)
			return
		case "add-food":
			err = AddFood()
		case "undo-food":
			err = UndoFood()
		case "update-food":
			err = UpdateFood()
		case "delete-food":
			err = DeleteFood()
		case "log":
			err = AddLog()
		case "undo-log":
			err = UndoLog()
		case "print-foods":
			err = PrintFoods()
		default:
			err = fmt.Errorf("unknown command: %s (run: go run . help)", os.Args[1])
		}
	}

	Report()

	for _, n := range notices {
		fmt.Println(n)
	}

	if err != nil {
		fmt.Printf("%serror: %v%s\n", Red, err, Reset)
	}
}
