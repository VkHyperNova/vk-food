package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tests in this file (each function is a group of named subtests):
//   TestInitBackupStorage   - the "Mount?" flow before a backup
//   TestDeviceInMounts      - is the drive listed in /proc/mounts
//   TestParseDeviceByLabel  - finding the drive in the lsblk output
//   TestDecodeLsblkLabel    - turning \x20 back into a space
//   TestSaveWithBackup      - saving locally, then to the USB drive
//   TestWriteJSON           - writing a JSON file
//
// Not tested, because they need the real drive, lsblk or a key press:
// IsDriveMounted, unlockAndMount, findDeviceByLabel, mountDevice, confirmMount.
//
// This file also uses the helpers fakeDrive, quiet and exists from log_test.go
// and capture from report_test.go.
//
// Run one group:  go test -v -run TestSaveWithBackup
// Run one test (spaces in the name become _):
//                   go test -v -run 'TestSaveWithBackup/skipped'
//
// The helper function is at the bottom of the file.

// ============================================================
// InitBackupStorage
// ============================================================

func TestInitBackupStorage(t *testing.T) {
	t.Run("already mounted: no question, no mounting", func(t *testing.T) {
		asked, mounted := false, false
		setDrive(t,
			func() (bool, error) { return true, nil },
			func() (bool, error) { asked = true; return true, nil },
			func() error { mounted = true; return nil })

		if err := InitBackupStorage(); err != nil {
			t.Fatal(err)
		}
		if asked || mounted {
			t.Errorf("asked = %v, mounted = %v, want both false", asked, mounted)
		}
	})

	t.Run("the mount check fails", func(t *testing.T) {
		setDrive(t,
			func() (bool, error) { return false, errors.New("boom") },
			func() (bool, error) { return true, nil },
			func() error { return nil })

		err := InitBackupStorage()
		if err == nil || errors.Is(err, ErrBackupSkipped) {
			t.Errorf("want a real error, got %v", err)
		}
	})

	t.Run("user says no: backup skipped", func(t *testing.T) {
		mounted := false
		setDrive(t,
			func() (bool, error) { return false, nil },
			func() (bool, error) { return false, nil },
			func() error { mounted = true; return nil })

		if err := InitBackupStorage(); !errors.Is(err, ErrBackupSkipped) {
			t.Errorf("got %v, want ErrBackupSkipped", err)
		}
		if mounted {
			t.Error("the drive must not be mounted after a No")
		}
	})

	t.Run("the question fails: backup skipped", func(t *testing.T) {
		setDrive(t,
			func() (bool, error) { return false, nil },
			func() (bool, error) { return true, errors.New("no terminal") },
			func() error { return nil })

		if err := InitBackupStorage(); !errors.Is(err, ErrBackupSkipped) {
			t.Errorf("got %v, want ErrBackupSkipped", err)
		}
	})

	t.Run("mounting fails", func(t *testing.T) {
		setDrive(t,
			func() (bool, error) { return false, nil },
			func() (bool, error) { return true, nil },
			func() error { return errors.New("udisksctl failed") })

		err := InitBackupStorage()
		if err == nil || errors.Is(err, ErrBackupSkipped) {
			t.Errorf("want a real error, got %v", err)
		}
	})

	t.Run("mounting works and the drive shows up", func(t *testing.T) {
		checks := 0
		setDrive(t,
			func() (bool, error) { checks++; return checks > 1, nil }, // no, then yes
			func() (bool, error) { return true, nil },
			func() error { return nil })

		if err := InitBackupStorage(); err != nil {
			t.Fatal(err)
		}
		if checks != 2 {
			t.Errorf("checked %d times, want 2 (before and after mounting)", checks)
		}
	})

	t.Run("mounting works but the drive still is not there", func(t *testing.T) {
		setDrive(t,
			func() (bool, error) { return false, nil },
			func() (bool, error) { return true, nil },
			func() error { return nil })

		err := InitBackupStorage()
		if err == nil || errors.Is(err, ErrBackupSkipped) {
			t.Errorf("want a real error, got %v", err)
		}
	})
}

// ============================================================
// deviceInMounts
// ============================================================

func TestDeviceInMounts(t *testing.T) {
	mounts := "sysfs /sys sysfs rw 0 0\n" +
		"/dev/sda2 / ext4 rw 0 0\n" +
		"/dev/sdb1 /media/veikko/VK\\040DATA exfat rw 0 0\n"

	t.Run("finds a mounted device", func(t *testing.T) {
		got, err := deviceInMounts(strings.NewReader(mounts), "/dev/sdb1")
		if err != nil || !got {
			t.Errorf("got %v, %v, want true", got, err)
		}
	})

	t.Run("device that is not mounted", func(t *testing.T) {
		got, err := deviceInMounts(strings.NewReader(mounts), "/dev/sdc1")
		if err != nil || got {
			t.Errorf("got %v, %v, want false", got, err)
		}
	})

	t.Run("a longer name does not match (sdb1 is not sdb10)", func(t *testing.T) {
		got, _ := deviceInMounts(strings.NewReader("/dev/sdb10 /mnt ext4 rw 0 0\n"), "/dev/sdb1")
		if got {
			t.Error("/dev/sdb1 must not match /dev/sdb10")
		}
	})

	t.Run("only the first column counts", func(t *testing.T) {
		got, _ := deviceInMounts(strings.NewReader("tmpfs /dev/sdb1 tmpfs rw 0 0\n"), "/dev/sdb1")
		if got {
			t.Error("a mount point named /dev/sdb1 is not the device")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		got, err := deviceInMounts(strings.NewReader(""), "/dev/sdb1")
		if err != nil || got {
			t.Errorf("got %v, %v, want false", got, err)
		}
	})
}

// ============================================================
// parseDeviceByLabel
// ============================================================

func TestParseDeviceByLabel(t *testing.T) {
	t.Run("finds the device", func(t *testing.T) {
		out := "sda\nsda1 EFI\nsda2 root\nsdb1 VK\\x20DATA\n"

		got, err := parseDeviceByLabel(out, "VK DATA")
		if err != nil || got != "/dev/sdb1" {
			t.Errorf("got %q, %v, want /dev/sdb1", got, err)
		}
	})

	t.Run("label without spaces", func(t *testing.T) {
		got, err := parseDeviceByLabel("sdb1 BACKUP\n", "BACKUP")
		if err != nil || got != "/dev/sdb1" {
			t.Errorf("got %q, %v", got, err)
		}
	})

	t.Run("skips devices without a label", func(t *testing.T) {
		got, err := parseDeviceByLabel("sda\nsdb\nsdc1 VK\\x20DATA\n", "VK DATA")
		if err != nil || got != "/dev/sdc1" {
			t.Errorf("got %q, %v, want /dev/sdc1", got, err)
		}
	})

	t.Run("drive not plugged in is an error naming the label", func(t *testing.T) {
		_, err := parseDeviceByLabel("sda1 EFI\nsda2 root\n", "VK DATA")
		if err == nil || !strings.Contains(err.Error(), "VK DATA") {
			t.Errorf("want an error that names the label, got %v", err)
		}
	})

	t.Run("empty output is an error", func(t *testing.T) {
		if _, err := parseDeviceByLabel("", "VK DATA"); err == nil {
			t.Error("expected an error")
		}
	})
}

// ============================================================
// decodeLsblkLabel
// ============================================================

func TestDecodeLsblkLabel(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"space", `VK\x20DATA`, "VK DATA"},
		{"tab", `A\x09B`, "A\tB"},
		{"newline", `A\x0aB`, "A\nB"},
		{"backslash", `A\x5cB`, `A\B`},
		{"plain text stays as it is", "BACKUP", "BACKUP"},
		{"several codes", `A\x20B\x20C`, "A B C"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := decodeLsblkLabel(tc.in); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// ============================================================
// SaveWithBackup
// ============================================================

func TestSaveWithBackup(t *testing.T) {
	data := []Food{{Name: "Oats"}}

	t.Run("drive mounted: local and backup are written", func(t *testing.T) {
		fakeDrive(t, true, nil, false)
		dir := t.TempDir()
		local := filepath.Join(dir, "local", "food.json")
		backup := filepath.Join(dir, "usb", "food.json")

		var err error
		out := capture(t, func() { err = SaveWithBackup(local, backup, data) })

		if err != nil {
			t.Fatal(err)
		}
		if !exists(local) || !exists(backup) {
			t.Errorf("local exists: %v, backup exists: %v, want both", exists(local), exists(backup))
		}
		if !strings.Contains(out, "Local save!") || !strings.Contains(out, "Backup save!") {
			t.Errorf("want both messages on screen:\n%s", out)
		}
	})

	t.Run("skipped: only local is written", func(t *testing.T) {
		fakeDrive(t, false, nil, false) // not mounted, user answers No
		dir := t.TempDir()
		local := filepath.Join(dir, "local", "food.json")
		backup := filepath.Join(dir, "usb", "food.json")

		var err error
		out := capture(t, func() { err = SaveWithBackup(local, backup, data) })

		if err != nil {
			t.Fatalf("a skipped backup is not an error, got %v", err)
		}
		if !exists(local) || exists(backup) {
			t.Errorf("local exists: %v, backup exists: %v, want true and false", exists(local), exists(backup))
		}
		if !strings.Contains(out, "Backup skipped") {
			t.Errorf("want the skipped message:\n%s", out)
		}
	})

	t.Run("the drive check fails: local is still saved", func(t *testing.T) {
		fakeDrive(t, false, errors.New("boom"), false)
		dir := t.TempDir()
		local := filepath.Join(dir, "food.json")

		var err error
		out := capture(t, func() { err = SaveWithBackup(local, filepath.Join(dir, "usb", "food.json"), data) })

		if err != nil {
			t.Fatalf("backup problems must not fail the save, got %v", err)
		}
		if !exists(local) {
			t.Error("local file should be saved")
		}
		if !strings.Contains(out, "Backup init failed") {
			t.Errorf("want the failure message:\n%s", out)
		}
	})

	t.Run("the backup write fails: local is still saved", func(t *testing.T) {
		fakeDrive(t, true, nil, false)
		dir := t.TempDir()
		local := filepath.Join(dir, "food.json")

		// a plain file where the backup folder should be, so the backup cannot be written
		blocker := filepath.Join(dir, "blocker")
		os.WriteFile(blocker, []byte("x"), 0o644)

		var err error
		out := capture(t, func() { err = SaveWithBackup(local, filepath.Join(blocker, "food.json"), data) })

		if err != nil {
			t.Fatalf("backup problems must not fail the save, got %v", err)
		}
		if !exists(local) {
			t.Error("local file should be saved")
		}
		if !strings.Contains(out, "Backup write failed") {
			t.Errorf("want the failure message:\n%s", out)
		}
	})

	t.Run("the local write fails: error, and no backup", func(t *testing.T) {
		fakeDrive(t, true, nil, false)
		quiet(t)
		dir := t.TempDir()

		blocker := filepath.Join(dir, "blocker")
		os.WriteFile(blocker, []byte("x"), 0o644)
		backup := filepath.Join(dir, "usb", "food.json")

		if err := SaveWithBackup(filepath.Join(blocker, "food.json"), backup, data); err == nil {
			t.Fatal("a failed local save must be returned")
		}
		if exists(backup) {
			t.Error("no backup should be written when the local save failed")
		}
	})
}

// ============================================================
// writeJSON
// ============================================================

func TestWriteJSON(t *testing.T) {
	t.Run("creates missing folders", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "a", "b", "food.json")

		if err := writeJSON(path, []Food{{Name: "Oats"}}); err != nil {
			t.Fatal(err)
		}
		if !exists(path) {
			t.Error("file was not created")
		}
	})

	t.Run("writes JSON that loads back", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "food.json")
		in := []Food{{Name: "Oats", Kcal: 370}, {Name: "Cheese", Salt: 1.8}}

		if err := writeJSON(path, in); err != nil {
			t.Fatal(err)
		}

		var out []Food
		if err := json.Unmarshal([]byte(readFile(t, path)), &out); err != nil {
			t.Fatal(err)
		}
		if len(out) != 2 || out[0] != in[0] || out[1] != in[1] {
			t.Errorf("got %+v, want %+v", out, in)
		}
	})

	t.Run("replaces an existing file completely", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "food.json")
		writeJSON(path, []Food{{Name: "A"}, {Name: "B"}, {Name: "C"}})

		if err := writeJSON(path, []Food{{Name: "D"}}); err != nil {
			t.Fatal(err)
		}

		var out []Food
		json.Unmarshal([]byte(readFile(t, path)), &out)
		if len(out) != 1 || out[0].Name != "D" {
			t.Errorf("got %+v, want only D", out)
		}
	})

	t.Run("a value that cannot be JSON is an error and writes nothing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "food.json")

		if err := writeJSON(path, make(chan int)); err == nil {
			t.Fatal("expected an error")
		}
		if exists(path) {
			t.Error("no file should be written")
		}
	})

	t.Run("a file where the folder should be is an error", func(t *testing.T) {
		dir := t.TempDir()
		blocker := filepath.Join(dir, "blocker")
		os.WriteFile(blocker, []byte("x"), 0o644)

		if err := writeJSON(filepath.Join(blocker, "food.json"), []Food{}); err == nil {
			t.Error("expected an error")
		}
	})
}

// ============================================================
// Helper (used by the tests above)
// ============================================================

// setDrive replaces the three drive functions with fakes for the length of the
// test: check (is it mounted?), ask (the "Mount?" question) and mount.
func setDrive(t *testing.T, check func() (bool, error), ask func() (bool, error), mount func() error) {
	t.Helper()

	oldCheck, oldAsk, oldMount := checkMounted, askMount, mountDrive
	t.Cleanup(func() { checkMounted, askMount, mountDrive = oldCheck, oldAsk, oldMount })

	checkMounted, askMount, mountDrive = check, ask, mount
}
