package main

import (
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	streakKcalLimit = 2200 // a day counts when it is under this many kcal
	streakDays      = 7    // days in a row for one "You did it! +1"
)

// Streak is the result of looking through the whole log.
type Streak struct {
	Current   int  // days in a row right now (ends today, or yesterday if today is still empty)
	Completed int  // how many times 7 days in a row were reached, in total
	JustDone  bool // today is the day that completed a 7-day run
}

// CalcStreak walks every day from the first log entry up to today.
// A day counts only if something was logged and the calories are under streakKcalLimit.
// Every full 7 good days in a row gives +1. After +1 the next run starts again from 0.
// An empty day in the past breaks the run. An empty today does not, because the day is not over.
func CalcStreak(foods []Food, entries []Log, today time.Time) Streak {
	var s Streak
	if len(entries) == 0 {
		return s
	}

	perDay := map[string][]Log{}
	first := entries[0].Time.Local()
	for _, e := range entries {
		key := dayKey(e.Time)
		perDay[key] = append(perDay[key], e)
		if e.Time.Before(first) {
			first = e.Time.Local()
		}
	}

	day := startOfDay(first)
	end := startOfDay(today)
	for ; !day.After(end); day = day.AddDate(0, 0, 1) {
		list := perDay[dayKey(day)]
		isToday := day.Equal(end)

		if len(list) == 0 {
			if !isToday {
				s.Current = 0
			}
			continue
		}

		if sumEntries(foods, list).Kcal < streakKcalLimit {
			s.Current++
			if s.Current%streakDays == 0 {
				s.Completed++
				s.JustDone = isToday
			}
		} else {
			s.Current = 0
		}
	}
	return s
}

func printStreak(out io.Writer, foods []Food, entries []Log, today time.Time) {
	s := CalcStreak(foods, entries, today)

	// progress inside the current run of 7: 0..7 (a finished run shows 7/7, not 0/7)
	filled := s.Current % streakDays
	if s.JustDone {
		filled = streakDays
	}

	fmt.Fprintf(out, "%sStreak%s %s%s%s %d/%d  (under %d kcal)\n",
		Bold, Reset,
		Green, strings.Repeat("●", filled)+Reset+strings.Repeat("○", streakDays-filled), Reset,
		filled, streakDays, streakKcalLimit)
	fmt.Fprintf(out, "7-day streaks done: %d\n", s.Completed)
	if s.JustDone {
		fmt.Fprintln(out, Bold+Green+"You did it! +1"+Reset)
	}
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Local().Date()
	return time.Date(y, m, d, 12, 0, 0, 0, time.Local) // noon, so daylight saving shifts cannot skip a day
}

func dayKey(t time.Time) string {
	return t.Local().Format("2006-01-02")
}