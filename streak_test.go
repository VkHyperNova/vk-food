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

	t.Run("7 good days in a row gives +1 on the 7th day", func(t *testing.T) {
		var logs []Log
		for n := 6; n >= 0; n-- {
			logs = append(logs, goodDay(n))
		}
		s := CalcStreak(reportFoods, logs, reportNow)
		if s.Completed != 1 || !s.JustDone {
			t.Errorf("got %+v, want Completed 1 and JustDone", s)
		}
	})

	t.Run("the 8th day is not 'just done' and starts the next run", func(t *testing.T) {
		var logs []Log
		for n := 7; n >= 0; n-- {
			logs = append(logs, goodDay(n))
		}
		s := CalcStreak(reportFoods, logs, reportNow)
		if s.Completed != 1 || s.JustDone || s.Current != 8 {
			t.Errorf("got %+v, want Completed 1, Current 8, not JustDone", s)
		}
	})

	t.Run("14 good days in a row gives +2", func(t *testing.T) {
		var logs []Log
		for n := 13; n >= 0; n-- {
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

		for _, want := range []string{"Streak ●●○○○○○ 2/7", "under 2200 kcal", "7-day streaks done: 0"} {
			if !strings.Contains(plain, want) {
				t.Errorf("missing %q:\n%s", want, plain)
			}
		}
		if strings.Contains(plain, "You did it") {
			t.Errorf("should not celebrate yet:\n%s", plain)
		}
	})

	t.Run("celebrates on the 7th day", func(t *testing.T) {
		var logs []Log
		for n := 6; n >= 0; n-- {
			logs = append(logs, goodDay(n))
		}
		var out strings.Builder
		printStreak(&out, reportFoods, logs, reportNow)
		plain := stripColour(out.String())

		for _, want := range []string{"●●●●●●● 7/7", "7-day streaks done: 1", "You did it! +1"} {
			if !strings.Contains(plain, want) {
				t.Errorf("missing %q:\n%s", want, plain)
			}
		}
	})
}