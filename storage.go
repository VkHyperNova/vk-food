package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/VkHyperNova/vkmount"
)

/* Storage */

var mountBackup = func() error {
	return vkmount.Drive{Label: driveLabel}.Mount()
}

func Save(local, backup string, source any) error {
	if err := writeJSON(local, source); err != nil {
		return err
	}

	fmt.Println(Green + "Local Saved!" + Reset)

	err := mountBackup()

	if errors.Is(err, vkmount.ErrSkipped) {
		return nil
	}

	if err != nil {
		return err
	}

	if err := writeJSON(backup, source); err != nil {
		return err
	}

	fmt.Println(Green + "Backup Saved!" + Reset)

	return nil
}

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
