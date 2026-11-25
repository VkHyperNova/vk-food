package main

import (
	"strings"
	"testing"
	"time"
)

// Tests in this file:
//   TestCalcStreak  - counting days in a row under 2200 kcal
//   TestPrintStreak - what the streak looks like on screen
//
// It uses reportNow and reportFoods from report_test.go.
// Oats has 370 kcal per 100 g: 500 g = 1850 kcal (good), 600 g = 2220 kcal (too much).

// goodDay / badDay make one log entry, n days before reportNow.
func goodDay(n int) Log { return Log{Time: reportNow.AddDate(0, 0, -n), Food: "Oats", Grams: 500} }
func badDay(n int) Log  { return Log{Time: reportNow.AddDate(0, 0, -n), Food: "Oats", Grams: 600} }

func TestCalcStreak(t *testing.T) {
	t.Run("no log is no streak", func(t *testing.T) {
		if s := CalcStreak(reportFoods, nil, reportNow); s != (Streak{}) {
			t.Errorf("got %+v, want zero", s)
		}
	})

	t.Run("counts days in a row up to today", func(t *testing.T) {
		s := CalcStreak(reportFoods, []Log{goodDay(2), goodDay(1), goodDay(0)}, reportNow)
		if s.Current != 3 || s.Completed != 0 || s.JustDone {
			t.Errorf("got %+v, want Current 3", s)
		}
	})

	t.Run("10 good days in a row gives +1 on the 10th day", func(t *testing.T) {
		var logs []Log
		for n := 9; n >= 0; n-- {
			logs = append(logs, goodDay(n))
		}
		s := CalcStreak(reportFoods, logs, reportNow)
		if s.Completed != 1 || !s.JustDone {
			t.Errorf("got %+v, want Completed 1 and JustDone", s)
		}
	})

	t.Run("the 11th day is not 'just done' and starts the next run", func(t *testing.T) {
		var logs []Log
		for n := 10; n >= 0; n-- {
			logs = append(logs, goodDay(n))
		}
		s := CalcStreak(reportFoods, logs, reportNow)
		if s.Completed != 1 || s.JustDone || s.Current != 11 {
			t.Errorf("got %+v, want Completed 1, Current 11, not JustDone", s)
		}
	})

	t.Run("20 good days in a row gives +2", func(t *testing.T) {
		var logs []Log
		for n := 19; n >= 0; n-- {
			logs = append(logs, goodDay(n))
		}
		if s := CalcStreak(reportFoods, logs, reportNow); s.Completed != 2 {
			t.Errorf("got %+v, want Completed 2", s)
		}
	})

	t.Run("a day over 2200 kcal breaks the run", func(t *testing.T) {
		logs := []Log{goodDay(3), goodDay(2), badDay(1), goodDay(0)}
		if s := CalcStreak(reportFoods, logs, reportNow); s.Current != 1 {
			t.Errorf("got %+v, want Current 1", s)
		}
	})

	t.Run("exactly 2200 kcal is not under the limit", func(t *testing.T) {
		// 2200 / 3.7 = 594.59 g would be too fiddly: use a food with 220 kcal per 100 g
		foods := []Food{{Name: "Test", Kcal: 220}}
		logs := []Log{{Time: reportNow, Food: "Test", Grams: 1000}}
		if s := CalcStreak(foods, logs, reportNow); s.Current != 0 {
			t.Errorf("got %+v, want Current 0", s)
		}
	})

	t.Run("an empty day in the past breaks the run", func(t *testing.T) {
		logs := []Log{goodDay(3), goodDay(1), goodDay(0)}
		if s := CalcStreak(reportFoods, logs, reportNow); s.Current != 2 {
			t.Errorf("got %+v, want Current 2", s)
		}
	})

	t.Run("an empty today keeps yesterday's run alive", func(t *testing.T) {
		logs := []Log{goodDay(2), goodDay(1)}
		if s := CalcStreak(reportFoods, logs, reportNow); s.Current != 2 || s.JustDone {
			t.Errorf("got %+v, want Current 2", s)
		}
	})

	t.Run("going over the limit today breaks the run", func(t *testing.T) {
		logs := []Log{goodDay(2), goodDay(1), badDay(0)}
		if s := CalcStreak(reportFoods, logs, reportNow); s.Current != 0 {
			t.Errorf("got %+v, want Current 0", s)
		}
	})

	t.Run("several entries on one day are added together", func(t *testing.T) {
		// 2 x 300 g oats = 2220 kcal
		logs := []Log{
			{Time: reportNow, Food: "Oats", Grams: 300},
			{Time: reportNow.Add(-time.Hour), Food: "Oats", Grams: 300},
		}
		if s := CalcStreak(reportFoods, logs, reportNow); s.Current != 0 {
			t.Errorf("got %+v, want Current 0", s)
		}
	})
}

func TestPrintStreak(t *testing.T) {
	t.Run("shows progress and the total", func(t *testing.T) {
		var out strings.Builder
		printStreak(&out, reportFoods, []Log{goodDay(1), goodDay(0)}, reportNow)
		plain := stripColour(out.String())

		for _, want := range []string{"Streak ○○○○○○○○●● 2/10", "under 2200 kcal", "10-day streaks done: 0"} {
			if !strings.Contains(plain, want) {
				t.Errorf("missing %q:\n%s", want, plain)
			}
		}
		if strings.Contains(plain, "You did it") {
			t.Errorf("should not celebrate yet:\n%s", plain)
		}
	})

	t.Run("colors recent days by calories", func(t *testing.T) {
		foods := []Food{
			{Name: "Green", Kcal: 220},
			{Name: "Yellow", Kcal: 222},
			{Name: "Red", Kcal: 320},
		}
		logs := []Log{
			{Time: reportNow.AddDate(0, 0, -2), Food: "Green", Grams: 1000},
			{Time: reportNow.AddDate(0, 0, -1), Food: "Yellow", Grams: 1000},
			{Time: reportNow, Food: "Red", Grams: 1000},
		}
		var out strings.Builder
		printStreak(&out, foods, logs, reportNow)

		wantDots := strings.Repeat("○", 7) +
			Bold + Green + "●" + Reset +
			Bold + Yellow + "●" + Reset +
			Bold + Red + "●" + Reset
		if !strings.Contains(out.String(), wantDots) {
			t.Errorf("dots = %q, want recent days colored by calories: %q", out.String(), wantDots)
		}
		if !strings.Contains(stripColour(out.String()), "0/10") {
			t.Errorf("streak progress should still count only under-limit days:\n%s", stripColour(out.String()))
		}
	})

	t.Run("celebrates on the 10th day", func(t *testing.T) {
		var logs []Log
		for n := 9; n >= 0; n-- {
			logs = append(logs, goodDay(n))
		}
		var out strings.Builder
		printStreak(&out, reportFoods, logs, reportNow)
		plain := stripColour(out.String())

		for _, want := range []string{"●●●●●●●●●● 10/10", "10-day streaks done: 1", "You did it! +1"} {
			if !strings.Contains(plain, want) {
				t.Errorf("missing %q:\n%s", want, plain)
			}
		}
	})
}
