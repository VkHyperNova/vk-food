package main

import (
	"math"
	"strings"
	"testing"
	"time"
)

// fixed "now" so the tests do not depend on the real date
var testNow = time.Date(2026, 10, 14, 12, 0, 0, 0, time.Local)

var testFoods = []Food{
	{Name: "Oats", Kcal: 370, Protein: 13, Carbs: 60, Fat: 7, Fiber: 10, Sugar: 1, Salt: 0.02},
	{Name: "Cheese", Kcal: 400, Protein: 25, Carbs: 1, Fat: 33, Fiber: 0, Sugar: 1, Salt: 1.8},
}

func near(a, b float64) bool { return math.Abs(a-b) < 0.001 }

func TestTotalsAdd(t *testing.T) {
	var tot Totals
	tot.Add(testFoods[0], 50) // half of 100 g
	tot.Add(testFoods[0], 50)

	if !near(tot.Kcal, 370) || !near(tot.Protein, 13) || !near(tot.Fiber, 10) {
		t.Errorf("unexpected totals: %+v", tot)
	}
}

func TestFindFood(t *testing.T) {
	f, ok := findFood(testFoods, "oATS")
	if !ok || f.Name != "Oats" {
		t.Errorf("findFood(oATS) = %v, %v", f, ok)
	}
	if _, ok := findFood(testFoods, "Pizza"); ok {
		t.Error("Pizza should not be found")
	}
}

func TestSumEntries(t *testing.T) {
	entries := []Log{
		{Time: testNow, Food: "Oats", Grams: 100},
		{Time: testNow, Food: "Cheese", Grams: 50},
		{Time: testNow, Food: "Pizza", Grams: 200},
	}
	tot, missing := sumEntries(testFoods, entries)

	if !near(tot.Kcal, 570) { // 370 + 200
		t.Errorf("Kcal = %v, want 570", tot.Kcal)
	}
	if !near(tot.Salt, 0.92) { // 0.02 + 0.9
		t.Errorf("Salt = %v, want 0.92", tot.Salt)
	}
	if len(missing) != 1 || missing[0] != "Pizza" {
		t.Errorf("missing = %v, want [Pizza]", missing)
	}
}

func TestLevelColor(t *testing.T) {
	// salt bands: green 1-5, yellow over 5, red over 10
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
}

func TestLogsOn(t *testing.T) {
	yesterday := testNow.AddDate(0, 0, -1)
	logs := []Log{
		{Time: testNow, Food: "Oats", Grams: 50},
		{Time: yesterday, Food: "Cheese", Grams: 30},
		{Time: testNow.Add(-3 * time.Hour), Food: "Oats", Grams: 20},
	}

	if got := LogsOn(logs, testNow); len(got) != 2 {
		t.Errorf("today: got %d entries, want 2", len(got))
	}
	if got := LogsOn(logs, yesterday); len(got) != 1 {
		t.Errorf("yesterday: got %d entries, want 1", len(got))
	}
}

func TestPrintWeek(t *testing.T) {
	logs := []Log{{Time: testNow.AddDate(0, 0, -2), Food: "Oats", Grams: 100}}

	var out strings.Builder
	printWeek(&out, testFoods, logs, testNow)
	s := out.String()

	// heading + header + 8 day rows (7 days back and today)
	if lines := strings.Split(strings.TrimSpace(s), "\n"); len(lines) != 10 {
		t.Fatalf("got %d lines, want 10:\n%s", len(lines), s)
	}
	if !strings.Contains(s, "Mon 12 Oct") || !strings.Contains(s, "370") {
		t.Errorf("logged day missing or wrong:\n%s", s)
	}
	if !strings.Contains(s, "Wed 14 Oct") {
		t.Errorf("today row missing:\n%s", s)
	}
	if !strings.Contains(s, " -") {
		t.Errorf("empty days should show -:\n%s", s)
	}
}

func TestPrintTodayNothingLogged(t *testing.T) {
	var out strings.Builder
	printToday(&out, testFoods, nil, testNow)

	if !strings.Contains(out.String(), "Nothing logged today.") {
		t.Errorf("output:\n%s", out.String())
	}
}

func TestPrintTodayTotals(t *testing.T) {
	logs := []Log{
		{Time: testNow, Food: "Oats", Grams: 100},
		{Time: testNow, Food: "Cheese", Grams: 100},
	}

	var out strings.Builder
	printToday(&out, testFoods, logs, testNow)
	s := out.String()

	for _, want := range []string{"Calories", "Protein", "Carbs", "Sugar", "Fat", "Fiber", "Salt", "770.00", "38.00", "1.82"} {
		if !strings.Contains(s, want) {
			t.Errorf("output missing %q:\n%s", want, s)
		}
	}
}

func TestPrintTodaySaltColour(t *testing.T) {
	tests := []struct {
		name  string
		grams float64 // grams of cheese (1.8 g salt per 100 g)
		want  string
	}{
		{"white below 1 g", 20, Reset},          // 0.36 g
		{"green at 1-5 g", 100, Bold + Green},   // 1.8 g
		{"yellow over 5 g", 400, Bold + Yellow}, // 7.2 g
		{"red over 10 g", 700, Bold + Red},      // 12.6 g
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logs := []Log{{Time: testNow, Food: "Cheese", Grams: tc.grams}}
			var out strings.Builder
			printToday(&out, testFoods, logs, testNow)

			var saltLine string
			for _, l := range strings.Split(out.String(), "\n") {
				if strings.Contains(l, "Salt") {
					saltLine = l
				}
			}
			if !strings.Contains(saltLine, "Salt      "+tc.want) {
				t.Errorf("salt line = %q, want colour %q right after the label", saltLine, tc.want)
			}
		})
	}
}

func TestPrintTodayMissingFood(t *testing.T) {
	logs := []Log{
		{Time: testNow, Food: "Oats", Grams: 100},
		{Time: testNow, Food: "Pizza", Grams: 200},
	}

	var out strings.Builder
	printToday(&out, testFoods, logs, testNow)
	s := out.String()

	if !strings.Contains(s, "Not in food list (skipped): Pizza") {
		t.Errorf("missing warning:\n%s", s)
	}
	if !strings.Contains(s, "?") {
		t.Errorf("row for unknown food should show ?:\n%s", s)
	}
}
