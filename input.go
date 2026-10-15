package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// writeJSON saves any value to a file as indented JSON.
func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func saveWithBackup(local, backup string, v any) error {
	if err := writeJSON(local, v); err != nil {
		return err
	}
	fmt.Println(Green + "Local save!" + Reset)

	if err := InitBackupStorage(); err != nil {
		if errors.Is(err, ErrBackupSkipped) {
			fmt.Println(Yellow + "Backup skipped" + Reset)
		} else {
			fmt.Println(Yellow + "Backup init failed: " + err.Error() + Reset)
		}
		return nil
	}
	if err := writeJSON(backup, v); err != nil {
		fmt.Println(Yellow + "Backup write failed: " + err.Error() + Reset)
		return nil
	}
	fmt.Println(Green + "Backup save!" + Reset)
	return nil
}
