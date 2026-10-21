package main

import (
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Tests in this file (each function is a group of named subtests):
//   TestTotalsAdd    - adding one food to the totals
//   TestFindFood     - looking up a food by name
//   TestSumEntries   - adding up many log entries
//   TestLevelColor   - white / green / yellow / red for a value
//   TestLevelsOf     - the limits of each nutrient (calories, protein, ... salt)
//   TestCell         - one coloured table cell
//   TestSameDay      - are two times on the same day
//   TestLogsOn       - log entries of one day
//   TestPrintWeek    - the "Last 7 days" table
//   TestPrintToday   - the "Today" table and the totals
//   TestReport       - the whole report on screen
//
// Not tested: clearScreen (it only runs the "clear" command).
// This file also uses the helpers useTempData and seedLogs from log_test.go.
//
// Run one group:  go test -v -run TestPrintToday
// Run one test (spaces in the name become _):
//                   go test -v -run 'TestPrintToday/salt'
//
// The helpers are at the bottom of the file.

// ============================================================
// Totals.add
// ============================================================

func TestTotalsAdd(t *testing.T) {
	t.Run("scales the values by the grams", func(t *testing.T) {
		var tot Totals
		tot.add(reportFoods[0], 50) // Oats, half of 100 g

		if !near(tot.Kcal, 185) || !near(tot.Protein, 6.5) || !near(tot.Fiber, 5) {
			t.Errorf("unexpected totals: %+v", tot)
		}
	})

	t.Run("adds up over several calls", func(t *testing.T) {
		var tot Totals
		tot.add(reportFoods[0], 50)
		tot.add(reportFoods[0], 50)

		if !near(tot.Kcal, 370) || !near(tot.Protein, 13) {
			t.Errorf("unexpected totals: %+v", tot)
		}
	})

	t.Run("zero grams adds nothing", func(t *testing.T) {
		var tot Totals
		tot.add(reportFoods[0], 0)

		if tot != (Totals{}) {
			t.Errorf("want all zeros, got %+v", tot)
		}
	})
}

// ============================================================
// findFood
// ============================================================

func TestFindFood(t *testing.T) {
	t.Run("finds a food", func(t *testing.T) {
		f, ok := findFood(reportFoods, "Oats")
		if !ok || f.Name != "Oats" {
			t.Errorf("got %v, %v", f, ok)
		}
	})

	t.Run("ignores upper and lower case", func(t *testing.T) {
		f, ok := findFood(reportFoods, "oATS")
		if !ok || f.Name != "Oats" {
			t.Errorf("got %v, %v", f, ok)
		}
	})

	t.Run("unknown food is not found", func(t *testing.T) {
		if _, ok := findFood(reportFoods, "Pizza"); ok {
			t.Error("Pizza should not be found")
		}
	})

	t.Run("empty food list", func(t *testing.T) {
		if _, ok := findFood(nil, "Oats"); ok {
			t.Error("nothing can be found in an empty list")
		}
	})
}

// ============================================================
// sumEntries
// ============================================================

func TestSumEntries(t *testing.T) {
	t.Run("adds up all entries", func(t *testing.T) {
		entries := []Log{
			{Time: reportNow, Food: "Oats", Grams: 100},
			{Time: reportNow, Food: "Cheese", Grams: 50},
		}
		tot := sumEntries(reportFoods, entries)

		if !near(tot.Kcal, 570) { // 370 + 200
			t.Errorf("Kcal = %v, want 570", tot.Kcal)
		}
		if !near(tot.Salt, 0.92) { // 0.02 + 0.9
			t.Errorf("Salt = %v, want 0.92", tot.Salt)
		}
	})

	t.Run("an unknown food adds nothing", func(t *testing.T) {
		entries := []Log{
			{Time: reportNow, Food: "Oats", Grams: 100},
			{Time: reportNow, Food: "Pizza", Grams: 200},
		}
		tot := sumEntries(reportFoods, entries)

		if !near(tot.Kcal, 370) {
			t.Errorf("Kcal = %v, want 370 (Pizza is not counted)", tot.Kcal)
		}
	})

	t.Run("no entries", func(t *testing.T) {
		tot := sumEntries(reportFoods, nil)

		if tot != (Totals{}) {
			t.Errorf("got %+v, want all zeros", tot)
		}
	})
}

// ============================================================
// levelColor
// ============================================================

func TestLevelColor(t *testing.T) {
	t.Run("salt: white, green, yellow, red", func(t *testing.T) {
		// salt bands: green from 1, yellow over 5, red over 10
		tests := []struct {
			v    float64
			want string
		}{
			{0.5, Reset},
			{1, Bold + Green},
			{5, Bold + Green},
			{5.1, Bold + Yellow},
			{10, Bold + Yellow},
			{10.1, Bold + Red},
		}
		for _, tc := range tests {
			if got := levelColor(tc.v, 1, 5, 10); got != tc.want {
				t.Errorf("levelColor(%v) = %q, want %q", tc.v, got, tc.want)
			}
		}
	})

	t.Run("green from 0 is green even for zero", func(t *testing.T) {
		if got := levelColor(0, 0, 25, 50); got != Bold+Green {
			t.Errorf("got %q, want green", got)
		}
	})
}

// ============================================================
// levelsOf
// ============================================================

// The limits of each nutrient: white below green, green up to the yellow limit,
// yellow up to the red limit, red above it. Change the numbers here when you
// change them in levelsOf.
func TestLevelsOf(t *testing.T) {
	tests := []struct {
		name               string
		green, yellow, red float64
		set                func(*Totals, float64)
		get                func(Levels) string
	}{
		{"calories", 0, 2400, 3100, func(t *Totals, v float64) { t.Kcal = v }, func(l Levels) string { return l.Kcal }},
		{"protein", 85, 150, 180, func(t *Totals, v float64) { t.Protein = v }, func(l Levels) string { return l.Protein }},
		{"carbs", 50, 150, 300, func(t *Totals, v float64) { t.Carbs = v }, func(l Levels) string { return l.Carbs }},
		{"fat", 45, 75, 90, func(t *Totals, v float64) { t.Fat = v }, func(l Levels) string { return l.Fat }},
		{"fiber", 25, 41, 50, func(t *Totals, v float64) { t.Fiber = v }, func(l Levels) string { return l.Fiber }},
		{"sugar", 0, 25, 50, func(t *Totals, v float64) { t.Sugar = v }, func(l Levels) string { return l.Sugar }},
		{"salt", 0, 4, 5, func(t *Totals, v float64) { t.Salt = v }, func(l Levels) string { return l.Salt }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			colourAt := func(v float64) string {
				var tot Totals
				tc.set(&tot, v)
				return tc.get(levelsOf(tot))
			}

			check := func(v float64, want string) {
				t.Helper()
				if got := colourAt(v); got != want {
					t.Errorf("%v = %q, want %q", v, got, want)
				}
			}

			if tc.green > 0 {
				check(tc.green-0.1, Reset) // just below green is white
			}
			check(tc.green, Bold+Green)
			check(tc.yellow, Bold+Green) // the yellow limit itself is still green
			check(tc.yellow+0.1, Bold+Yellow)
			check(tc.red, Bold+Yellow) // the red limit itself is still yellow
			check(tc.red+0.1, Bold+Red)
		})
	}

	t.Run("nothing eaten", func(t *testing.T) {
		got := levelsOf(Totals{})
		// the nutrients that turn green from 0 are green, the others white
		if got.Kcal != Bold+Green || got.Sugar != Bold+Green || got.Salt != Bold+Green {
			t.Errorf("calories, sugar and salt start green at 0: %+v", got)
		}
		if got.Protein != Reset || got.Carbs != Reset || got.Fat != Reset || got.Fiber != Reset {
			t.Errorf("protein, carbs, fat and fiber start white: %+v", got)
		}
	})
}

// ============================================================
// cell
// ============================================================

func TestCell(t *testing.T) {
	t.Run("pads to 9 characters before the colour", func(t *testing.T) {
		want := Bold + Green + "       12" + Reset
		if got := cell(Bold+Green, "12"); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("visible width is 9 without the colour codes", func(t *testing.T) {
		for _, text := range []string{"-", "0", "1234", "203.0"} {
			if got := stripColour(cell(Bold+Red, text)); len(got) != 9 {
				t.Errorf("cell(%q) is %d wide, want 9: %q", text, len(got), got)
			}
		}
	})
}

// ============================================================
// sameDay
// ============================================================

func TestSameDay(t *testing.T) {
	t.Run("same day, different hours", func(t *testing.T) {
		morning := time.Date(2026, 10, 14, 0, 5, 0, 0, time.Local)
		night := time.Date(2026, 10, 14, 23, 55, 0, 0, time.Local)

		if !sameDay(morning, night) {
			t.Error("00:05 and 23:55 are on the same day")
		}
	})

	t.Run("next day is different", func(t *testing.T) {
		night := time.Date(2026, 10, 14, 23, 59, 0, 0, time.Local)
		after := time.Date(2026, 10, 15, 0, 1, 0, 0, time.Local)

		if sameDay(night, after) {
			t.Error("23:59 and 00:01 the next day are different days")
		}
	})

	t.Run("same date in a different year is different", func(t *testing.T) {
		a := time.Date(2026, 10, 14, 12, 0, 0, 0, time.Local)
		b := time.Date(2025, 10, 14, 12, 0, 0, 0, time.Local)

		if sameDay(a, b) {
			t.Error("14 Oct 2026 and 14 Oct 2025 are different days")
		}
	})
}

// ============================================================
// logsOn
// ============================================================

func TestLogsOn(t *testing.T) {
	t.Run("returns only the entries of that day", func(t *testing.T) {
		yesterday := reportNow.AddDate(0, 0, -1)
		logs := []Log{
			{Time: reportNow, Food: "Oats", Grams: 50},
			{Time: yesterday, Food: "Cheese", Grams: 30},
			{Time: reportNow.Add(-3 * time.Hour), Food: "Oats", Grams: 20},
		}

		if got := logsOn(logs, reportNow); len(got) != 2 {
			t.Errorf("today: got %d entries, want 2", len(got))
		}
		if got := logsOn(logs, yesterday); len(got) != 1 {
			t.Errorf("yesterday: got %d entries, want 1", len(got))
		}
	})

	t.Run("a day without entries", func(t *testing.T) {
		logs := []Log{{Time: reportNow, Food: "Oats", Grams: 50}}

		if got := logsOn(logs, reportNow.AddDate(0, 0, -5)); len(got) != 0 {
			t.Errorf("got %d entries, want 0", len(got))
		}
	})
}

// ============================================================
// printWeek
// ============================================================

func TestPrintWeek(t *testing.T) {
	logs := []Log{{Time: reportNow.AddDate(0, 0, -2), Food: "Oats", Grams: 100}}

	var out strings.Builder
	printWeek(&out, reportFoods, logs, reportNow)
	s := out.String()

	t.Run("heading, header and 8 day rows", func(t *testing.T) {
		// 7 days back and today
		if lines := strings.Split(strings.TrimSpace(s), "\n"); len(lines) != 10 {
			t.Errorf("got %d lines, want 10:\n%s", len(lines), s)
		}
	})

	t.Run("a logged day shows its totals", func(t *testing.T) {
		if !strings.Contains(s, "Mon 12 Oct") || !strings.Contains(s, "370") {
			t.Errorf("logged day missing or wrong:\n%s", s)
		}
	})

	t.Run("today is the last row", func(t *testing.T) {
		if !strings.Contains(s, "Wed 14 Oct") {
			t.Errorf("today row missing:\n%s", s)
		}
	})

	t.Run("days without entries show a dash", func(t *testing.T) {
		if !strings.Contains(s, " -") {
			t.Errorf("empty days should show -:\n%s", s)
		}
	})

	t.Run("days without entries have no colour", func(t *testing.T) {
		for _, line := range strings.Split(s, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "Wed 07 Oct") && strings.Contains(line, "\033[") {
				t.Errorf("an empty day should not be coloured: %q", line)
			}
		}
	})

	t.Run("columns line up when the colours are removed", func(t *testing.T) {
		// 10 for the date + 7 columns of 9
		for _, line := range strings.Split(strings.TrimSpace(stripColour(s)), "\n")[1:] {
			if len(line) != 10+7*9 {
				t.Errorf("row is %d wide, want %d: %q", len(line), 10+7*9, line)
			}
		}
	})

	t.Run("numbers are coloured like the limits", func(t *testing.T) {
		// 1000 g of Cheese in one day: 4000 kcal, 250 g protein, 10 g carbs,
		// 330 g fat, 0 g fiber, 10 g sugar, 18 g salt
		logs := []Log{{Time: reportNow, Food: "Cheese", Grams: 1000}}

		var out strings.Builder
		printWeek(&out, reportFoods, logs, reportNow)

		var today string
		for _, line := range strings.Split(out.String(), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "Wed 14 Oct") {
				today = line
			}
		}
		for _, want := range []string{
			Bold + Red + "     4000" + Reset,   // calories: red
			Bold + Red + "    250.0" + Reset,   // protein: red
			Reset + "     10.0" + Reset,        // carbs: white (below 50)
			Bold + Red + "    330.0" + Reset,   // fat: red
			Reset + "      0.0" + Reset,        // fiber: white (below 25)
			Bold + Green + "     10.0" + Reset, // sugar: green (from 0)
			Bold + Red + "     18.0" + Reset,   // salt: red
		} {
			if !strings.Contains(today, want) {
				t.Errorf("today row is missing %q:\n%q", want, today)
			}
		}
	})
}

// ============================================================
// printToday
// ============================================================

func TestPrintToday(t *testing.T) {
	t.Run("nothing logged", func(t *testing.T) {
		var out strings.Builder
		printToday(&out, reportFoods, nil, reportNow)

		if !strings.Contains(out.String(), "Nothing logged today.") {
			t.Errorf("output:\n%s", out.String())
		}
	})

	t.Run("shows a row for each entry", func(t *testing.T) {
		at := time.Date(2026, 10, 14, 8, 30, 0, 0, time.Local)
		logs := []Log{{Time: at, Food: "Oats", Grams: 50}}

		var out strings.Builder
		printToday(&out, reportFoods, logs, reportNow)
		s := out.String()

		for _, want := range []string{"08:30", "Oats", "50", "185"} { // 185 kcal
			if !strings.Contains(s, want) {
				t.Errorf("output missing %q:\n%s", want, s)
			}
		}
	})

	t.Run("shows all the totals", func(t *testing.T) {
		logs := []Log{
			{Time: reportNow, Food: "Oats", Grams: 100},
			{Time: reportNow, Food: "Cheese", Grams: 100},
		}

		var out strings.Builder
		printToday(&out, reportFoods, logs, reportNow)
		s := out.String()

		labels := []string{"Calories", "Protein", "Carbs", "Sugar", "Fat", "Fiber", "Salt"}
		for _, want := range append(labels, "38.00", "1.82") { // protein 13+25, salt 0.02+1.8
			if !strings.Contains(s, want) {
				t.Errorf("output missing %q:\n%s", want, s)
			}
		}
	})

	t.Run("salt colour follows the limits", func(t *testing.T) {
		tests := []struct {
			name  string
			grams float64 // grams of Cheese, 1.8 g salt per 100 g
			want  string
		}{
			{"green up to 4 g", 100, Bold + Green},  // 1.8 g
			{"yellow over 4 g", 250, Bold + Yellow}, // 4.5 g
			{"red over 5 g", 400, Bold + Red},       // 7.2 g
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				logs := []Log{{Time: reportNow, Food: "Cheese", Grams: tc.grams}}
				var out strings.Builder
				printToday(&out, reportFoods, logs, reportNow)

				var saltLine string
				for _, l := range strings.Split(out.String(), "\n") {
					if strings.Contains(l, "Salt") {
						saltLine = l
					}
				}
				// the text after the label "Salt" and its spaces must start with the colour
				afterLabel := strings.TrimLeft(strings.TrimPrefix(strings.TrimSpace(saltLine), "Salt"), " ")
				if !strings.HasPrefix(afterLabel, tc.want) {
					t.Errorf("salt line = %q, want colour %q right after the label", saltLine, tc.want)
				}
			})
		}
	})

	t.Run("shows the limit of each total", func(t *testing.T) {
		logs := []Log{{Time: reportNow, Food: "Oats", Grams: 100}}
		var out strings.Builder
		printToday(&out, reportFoods, logs, reportNow)
		s := out.String()

		for _, want := range []string{"(3100)", "(180)", "(300)", "(90)", "(50)", "(5)"} {
			if !strings.Contains(s, want) {
				t.Errorf("output missing the limit %q:\n%s", want, s)
			}
		}
	})

	t.Run("unknown food gets a ? and is not counted", func(t *testing.T) {
		logs := []Log{
			{Time: reportNow, Food: "Oats", Grams: 100},
			{Time: reportNow, Food: "Pizza", Grams: 200},
		}

		var out strings.Builder
		printToday(&out, reportFoods, logs, reportNow)
		s := out.String()

		if !strings.Contains(s, "Pizza") || !strings.Contains(s, "?") {
			t.Errorf("the Pizza row should show a ?:\n%s", s)
		}
		if !strings.Contains(s, "13.00") { // protein of the Oats only
			t.Errorf("Pizza should add nothing to the totals:\n%s", s)
		}
	})
}

// ============================================================
// Report
// ============================================================

func TestReport(t *testing.T) {
	t.Run("shows foods, last 7 days and today", func(t *testing.T) {
		useTempData(t) // saves Oats, Cheese, Peanut butter
		seedLogs(t)    // three entries logged now

		out := captureReport(t)

		for _, want := range []string{"Last 7 days", "Today,", "Peanut butter", "Calories"} {
			if !strings.Contains(out, want) {
				t.Errorf("output missing %q:\n%s", want, out)
			}
		}
	})

	t.Run("works on a first run with no files", func(t *testing.T) {
		useTempData(t)
		os.Remove(Foods)

		out := captureReport(t)

		if !strings.Contains(out, "Nothing logged today.") {
			t.Errorf("output:\n%s", out)
		}
	})

	t.Run("a broken food file shows its error but not a crash", func(t *testing.T) {
		useTempData(t)
		os.WriteFile(Foods, []byte("not json"), 0o644)

		out := captureReport(t)

		if !strings.Contains(out, "LoadFood:") {
			t.Errorf("expected the LoadFood error on screen:\n%s", out)
		}
	})
}

// ============================================================
// Helpers (used by the tests above)
// ============================================================

// A fixed "today" (Wednesday 14 Oct 2026), so the tests do not depend on the real date.
var reportNow = time.Date(2026, 10, 14, 12, 0, 0, 0, time.Local)

var reportFoods = []Food{
	{Name: "Oats", Kcal: 370, Protein: 13, Carbs: 60, Fat: 7, Fiber: 10, Sugar: 1, Salt: 0.02},
	{Name: "Cheese", Kcal: 400, Protein: 25, Carbs: 1, Fat: 33, Fiber: 0, Sugar: 1, Salt: 1.8},
}

// near compares numbers with a tiny margin, because of rounding in decimals.
func near(a, b float64) bool {
	d := a - b
	return d < 0.001 && d > -0.001
}

// capture runs fn and returns everything it printed to the terminal.
func capture(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = old })

	done := make(chan string)
	go func() {
		data, _ := io.ReadAll(r)
		done <- string(data)
	}()

	fn()

	w.Close()
	os.Stdout = old
	return <-done
}

// stripColour removes the colour codes (ESC [ ... m) so the text can be measured.
func stripColour(s string) string {
	return regexp.MustCompile("\x1b\\[[0-9;]*m").ReplaceAllString(s, "")
}

// captureReport runs Report() and returns what it printed.
func captureReport(t *testing.T) string {
	t.Helper()

	t.Setenv("TERM", "xterm") // ClearScreen runs "clear", which needs TERM
	return capture(t, Report)
}
