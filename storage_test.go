package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/VkHyperNova/vkmount"
)

// Tests in this file (each function is a group of named subtests):
//   TestSave        - saving locally, then to the USB drive
//   TestWriteJSON   - writing a JSON file
//
// Mounting the drive and the Yes/No question live in the vkmount package and
// are tested there. Here the mounting is replaced by fakeMount (log_test.go),
// so no real drive and no question on screen are needed.
//
// This file also uses the helpers fakeMount, quiet, exists and readFile from
// log_test.go and capture from report_test.go.
//
// Run one group:  go test -v -run TestSave
// Run one test (spaces in the name become _):
//                   go test -v -run 'TestSave/the_user_says_no'

// ============================================================
// Save
// ============================================================

func TestSave(t *testing.T) {
	foods := []Food{{Name: "Oats", Kcal: 370}}

	t.Run("drive ready: saved locally and on the drive", func(t *testing.T) {
		dir := t.TempDir()
		local, backup := filepath.Join(dir, "local.json"), filepath.Join(dir, "usb", "backup.json")
		fakeMount(t, nil)

		if err := Save(local, backup, foods); err != nil {
			t.Fatal(err)
		}

		if !exists(local) || !exists(backup) {
			t.Fatalf("local exists = %v, backup exists = %v, want both", exists(local), exists(backup))
		}
		if readFile(t, local) != readFile(t, backup) {
			t.Error("the backup must have the same content as the local file")
		}
	})

	t.Run("the user says no: only the local file", func(t *testing.T) {
		dir := t.TempDir()
		local, backup := filepath.Join(dir, "local.json"), filepath.Join(dir, "usb", "backup.json")
		fakeMount(t, vkmount.ErrSkipped)

		if err := Save(local, backup, foods); err != nil {
			t.Fatalf("a skipped backup is not an error: %v", err)
		}

		if !exists(local) {
			t.Error("the local file must be saved")
		}
		if exists(backup) {
			t.Error("nothing may be written to the drive after No")
		}
	})

	t.Run("mounting fails: error, but the local file is saved", func(t *testing.T) {
		dir := t.TempDir()
		local, backup := filepath.Join(dir, "local.json"), filepath.Join(dir, "usb", "backup.json")
		boom := errors.New("target is busy")
		fakeMount(t, boom)

		err := Save(local, backup, foods)
		if !errors.Is(err, boom) {
			t.Fatalf("err = %v, want %v", err, boom)
		}

		if !exists(local) {
			t.Error("the local file must be saved before the drive is tried")
		}
		if exists(backup) {
			t.Error("nothing may be written to the drive when mounting failed")
		}
	})

	t.Run("writing the backup fails: error, but the local file is saved", func(t *testing.T) {
		dir := t.TempDir()
		local := filepath.Join(dir, "local.json")
		blocker := filepath.Join(dir, "blocker")
		os.WriteFile(blocker, []byte("x"), 0o644)
		backup := filepath.Join(blocker, "backup.json") // a file where the folder should be
		fakeMount(t, nil)

		if err := Save(local, backup, foods); err == nil {
			t.Fatal("expected an error")
		}
		if !exists(local) {
			t.Error("the local file must be saved")
		}
	})

	t.Run("writing the local file fails: error, the drive is not touched", func(t *testing.T) {
		dir := t.TempDir()
		blocker := filepath.Join(dir, "blocker")
		os.WriteFile(blocker, []byte("x"), 0o644)
		local := filepath.Join(blocker, "local.json") // a file where the folder should be
		backup := filepath.Join(dir, "usb", "backup.json")
		calls := fakeMount(t, nil)

		if err := Save(local, backup, foods); err == nil {
			t.Fatal("expected an error")
		}
		if *calls != 0 {
			t.Errorf("mounting was asked for %d times, want 0 (nothing to back up)", *calls)
		}
		if exists(backup) {
			t.Error("no backup without a local save")
		}
	})

	t.Run("asks to mount once per save", func(t *testing.T) {
		dir := t.TempDir()
		calls := fakeMount(t, nil)

		Save(filepath.Join(dir, "a.json"), filepath.Join(dir, "usb", "a.json"), foods)
		if *calls != 1 {
			t.Errorf("mounting was asked for %d times, want 1", *calls)
		}
	})

	t.Run("a second save replaces both files", func(t *testing.T) {
		dir := t.TempDir()
		local, backup := filepath.Join(dir, "local.json"), filepath.Join(dir, "usb", "backup.json")
		fakeMount(t, nil)

		Save(local, backup, []Food{{Name: "A"}, {Name: "B"}})
		if err := Save(local, backup, []Food{{Name: "C"}}); err != nil {
			t.Fatal(err)
		}

		for _, path := range []string{local, backup} {
			var out []Food
			json.Unmarshal([]byte(readFile(t, path)), &out)
			if len(out) != 1 || out[0].Name != "C" {
				t.Errorf("%s = %+v, want only C", path, out)
			}
		}
	})

	t.Run("prints what happened", func(t *testing.T) {
		tests := []struct {
			name  string
			mount error
			want  []string
			not   []string
		}{
			{"drive ready", nil, []string{"Local Saved!", "Backup Saved!"}, []string{"skipped"}},
			{"user says no", vkmount.ErrSkipped, []string{"Local Saved!"}, []string{"Backup Saved!"}},
			{"mounting fails", errors.New("boom"), []string{"Local Saved!"}, []string{"Backup Saved!", "skipped"}},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				dir := t.TempDir()
				fakeMount(t, tc.mount)

				out := capture(t, func() {
					Save(filepath.Join(dir, "local.json"), filepath.Join(dir, "usb", "backup.json"), foods)
				})

				for _, want := range tc.want {
					if !strings.Contains(out, want) {
						t.Errorf("output missing %q:\n%s", want, out)
					}
				}
				for _, not := range tc.not {
					if strings.Contains(out, not) {
						t.Errorf("output should not contain %q:\n%s", not, out)
					}
				}
			})
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
