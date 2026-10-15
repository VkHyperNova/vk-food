package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

/* Pure helpers */

func TestDecodeLsblkLabel(t *testing.T) {
	if got := decodeLsblkLabel(`VK\x20DATA`); got != "VK DATA" {
		t.Errorf("got %q, want %q", got, "VK DATA")
	}
}

// Sample output of: lsblk -o NAME,LABEL -r -n
const sampleLsblk = "sda \n" +
	"sda1 ESP\n" +
	"sda2 \n" +
	"sdb \n" +
	`sdb1 VK\x20DATA` + "\n"

func TestParseDeviceByLabel(t *testing.T) {
	got, err := parseDeviceByLabel(sampleLsblk, "VK DATA")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/dev/sdb1" {
		t.Errorf("got %q, want /dev/sdb1", got)
	}
}

func TestParseDeviceByLabelNotPluggedIn(t *testing.T) {
	if _, err := parseDeviceByLabel(sampleLsblk, "OTHER DRIVE"); err == nil {
		t.Error("expected an error when the label is not present")
	}
}

// Sample content of /proc/mounts (spaces in paths are written as \040)
const sampleMounts = "/dev/sda2 / ext4 rw,relatime 0 0\n" +
	`/dev/sdb1 /media/veikko/VK\040DATA exfat rw,nosuid 0 0` + "\n"

func TestDeviceInMounts(t *testing.T) {
	mounted, err := deviceInMounts(strings.NewReader(sampleMounts), "/dev/sdb1")
	if err != nil || !mounted {
		t.Errorf("sdb1 should be mounted, got %v, %v", mounted, err)
	}
}

func TestDeviceNotInMounts(t *testing.T) {
	// /dev/sdb is a prefix of /dev/sdb1 but is a different device
	mounted, err := deviceInMounts(strings.NewReader(sampleMounts), "/dev/sdb")
	if err != nil || mounted {
		t.Errorf("sdb should not count as mounted, got %v, %v", mounted, err)
	}
}

func TestPaths(t *testing.T) {
	if filepath.IsAbs(FoodFile) || filepath.IsAbs(LogFile) {
		t.Errorf("local files should be relative: %s, %s", FoodFile, LogFile)
	}
	for _, p := range []string{FoodFileBackup, LogFileBackup} {
		if !strings.HasPrefix(p, driveMountPoint+"/") {
			t.Errorf("backup file %s should be under %s", p, driveMountPoint)
		}
	}
	if FoodFile == LogFile || FoodFileBackup == LogFileBackup {
		t.Error("food and log must use different files")
	}
}

/* InitBackupStorage, with the drive and the prompt replaced by fakes */

type fakes struct {
	mountedResults []bool // successive answers of the "is it mounted?" check
	mountedErr     error
	userAnswer     bool
	userErr        error
	mountErr       error

	checks, asked, mountCalls int
}

func installFakes(t *testing.T, f *fakes) {
	t.Helper()
	oldCheck, oldAsk, oldMount := checkMounted, askMount, mountDrive
	t.Cleanup(func() { checkMounted, askMount, mountDrive = oldCheck, oldAsk, oldMount })

	checkMounted = func() (bool, error) {
		i := f.checks
		f.checks++
		if f.mountedErr != nil {
			return false, f.mountedErr
		}
		if i >= len(f.mountedResults) {
			i = len(f.mountedResults) - 1
		}
		return f.mountedResults[i], nil
	}
	askMount = func() (bool, error) {
		f.asked++
		return f.userAnswer, f.userErr
	}
	mountDrive = func() error {
		f.mountCalls++
		return f.mountErr
	}
}

func TestInitBackupAlreadyMounted(t *testing.T) {
	f := &fakes{mountedResults: []bool{true}}
	installFakes(t, f)

	if err := InitBackupStorage(); err != nil {
		t.Fatal(err)
	}
	if f.asked != 0 || f.mountCalls != 0 {
		t.Errorf("should not ask or mount, asked=%d mounted=%d", f.asked, f.mountCalls)
	}
}

func TestInitBackupUserDeclines(t *testing.T) {
	f := &fakes{mountedResults: []bool{false}, userAnswer: false}
	installFakes(t, f)

	err := InitBackupStorage()
	if !errors.Is(err, ErrBackupSkipped) {
		t.Errorf("expected ErrBackupSkipped, got %v", err)
	}
	if f.mountCalls != 0 {
		t.Error("must not mount when the user says no")
	}
}

func TestInitBackupPromptFails(t *testing.T) {
	f := &fakes{mountedResults: []bool{false}, userErr: errors.New("ctrl+c")}
	installFakes(t, f)

	err := InitBackupStorage()
	if !errors.Is(err, ErrBackupSkipped) {
		t.Errorf("expected ErrBackupSkipped, got %v", err)
	}
	if f.mountCalls != 0 {
		t.Error("must not mount when the prompt fails")
	}
}

func TestInitBackupMountsWhenUserAgrees(t *testing.T) {
	f := &fakes{mountedResults: []bool{false, true}, userAnswer: true}
	installFakes(t, f)

	if err := InitBackupStorage(); err != nil {
		t.Fatal(err)
	}
	if f.mountCalls != 1 {
		t.Errorf("expected exactly one mount, got %d", f.mountCalls)
	}
}

func TestInitBackupMountFails(t *testing.T) {
	boom := errors.New("udisksctl failed")
	f := &fakes{mountedResults: []bool{false}, userAnswer: true, mountErr: boom}
	installFakes(t, f)

	err := InitBackupStorage()
	if err == nil || errors.Is(err, ErrBackupSkipped) {
		t.Fatalf("expected a mount error, got %v", err)
	}
	if !errors.Is(err, boom) {
		t.Errorf("the original error should be wrapped, got %v", err)
	}
}

func TestInitBackupStillNotMountedAfterMount(t *testing.T) {
	f := &fakes{mountedResults: []bool{false, false}, userAnswer: true}
	installFakes(t, f)

	err := InitBackupStorage()
	if err == nil || !strings.Contains(err.Error(), "still not mounted") {
		t.Errorf("expected a 'still not mounted' error, got %v", err)
	}
}

func TestInitBackupMountCheckFails(t *testing.T) {
	f := &fakes{mountedErr: errors.New("lsblk missing")}
	installFakes(t, f)

	err := InitBackupStorage()
	if err == nil || errors.Is(err, ErrBackupSkipped) {
		t.Fatalf("expected a check error, got %v", err)
	}
	if f.asked != 0 {
		t.Error("should not ask when the check itself failed")
	}
}
