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
var FoodFileBackup = filepath.Join(driveMountPoint, BaseLocal, BaseDB, "food.json")
var LogFileBackup = filepath.Join(driveMountPoint, BaseLocal, BaseDB, "log.json")

var ErrBackupSkipped = errors.New("backup skipped")

func main() {
	var err error

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "add-food":
			err = AddFood()
		case "undofood":
			err = UndoFood()
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
