package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/huh"
)

/* Storage */

func Save(local, backup string, v any) error {
	if err := writeJSON(local, v); err != nil {
		return err
	}

	fmt.Println(Green + "Local Saved!" + Reset)

	SaveBackup(backup, v)

	return nil
}

func SaveBackup(destination string, v any) error {
	if err := initBackupStorage(); err != nil {
		if errors.Is(err, ErrBackupSkipped) {
			fmt.Println(Yellow + "Backup skipped" + Reset)
		} else {
			fmt.Println(Yellow + "Backup init failed: " + err.Error() + Reset)
		}
		return nil
	}
	if err := writeJSON(destination, v); err != nil {
		fmt.Println(Yellow + "Backup write failed: " + err.Error() + Reset)
		return nil
	}
	fmt.Println(Green + "Backup Saved!" + Reset)

	return nil
}

/* Mounting */

// These are variables so that tests can replace them with fakes.
var (
	checkMounted = isDriveMounted
	askMount     = confirmMount
	mountDrive   = unlockAndMount
)

func initBackupStorage() error {
	mounted, err := checkMounted()
	if err != nil {
		return fmt.Errorf("mount check failed: %w", err)
	}

	if mounted {
		return nil // already mounted, nothing to do
	}

	mount, err := askMount()
	if err != nil || !mount {
		return ErrBackupSkipped
	}

	if err := mountDrive(); err != nil {
		return fmt.Errorf("failed to mount drive: %w", err)
	}

	mounted, err = checkMounted()
	if err != nil {
		return fmt.Errorf("mount check failed: %w", err)
	}
	if !mounted {
		return fmt.Errorf("drive still not mounted after mount attempt")
	}
	return nil
}

func isDriveMounted() (bool, error) {
	if runtime.GOOS != "linux" {
		return false, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	device, err := findDeviceByLabel(driveLabel)
	if err != nil {
		return false, fmt.Errorf("could not resolve device: %w", err)
	}

	file, err := os.Open("/proc/mounts")
	if err != nil {
		return false, fmt.Errorf("cannot open /proc/mounts: %w", err)
	}
	defer file.Close()

	return deviceInMounts(file, device)
}

func deviceInMounts(mounts io.Reader, device string) (bool, error) {
	scanner := bufio.NewScanner(mounts)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " ", 3)
		if len(parts) >= 2 && parts[0] == device {
			return true, nil
		}
	}
	return false, scanner.Err()
}

func unlockAndMount() error {
	device, err := findDeviceByLabel(driveLabel)
	if err != nil {
		return fmt.Errorf("could not find drive: %w", err)
	}
	fmt.Printf("Found drive at %s\n", device)

	if err := mountDevice(device); err != nil {
		return fmt.Errorf("mount failed: %w", err)
	}
	fmt.Printf("Drive mounted at %s\n", driveMountPoint)
	return nil
}

func findDeviceByLabel(label string) (string, error) {
	out, err := exec.Command("lsblk", "-o", "NAME,LABEL", "-r", "-n").Output()
	if err != nil {
		return "", fmt.Errorf("lsblk failed: %w", err)
	}
	return parseDeviceByLabel(string(out), label)
}

func parseDeviceByLabel(lsblkOutput, label string) (string, error) {
	for _, line := range strings.Split(lsblkOutput, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && decodeLsblkLabel(fields[1]) == label {
			return "/dev/" + fields[0], nil
		}
	}
	return "", fmt.Errorf("label '%s' not found (is the drive plugged in?)", label)
}

func decodeLsblkLabel(s string) string {
	return strings.NewReplacer(
		`\x20`, " ",
		`\x09`, "\t",
		`\x0a`, "\n",
		`\x5c`, `\`,
	).Replace(s)
}

func mountDevice(device string) error {
	cmd := exec.Command("udisksctl", "mount", "-b", device)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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

func confirmMount() (bool, error) {
	mount := true // Yes is preselected, so Enter mounts
	err := huh.NewConfirm().
		Title("Mount?").
		Affirmative("Yes").
		Negative("No").
		Value(&mount).
		Run()
	return mount, err
}
