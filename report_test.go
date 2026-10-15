package main

import (
	"bytes"
	"math"
	"strings"
	"testing"
	"time"
)

func approx(a, b float64) bool { return math.Abs(a-b) < 0.001 }

// A fixed "today" so the tests give the same result on every day.
var testNow = time.Date(2026, 10, 14, 15, 0, 0, 0, time.Local)

// at returns noon, daysAgo days before testNow.
func at(daysAgo int) time.Time {
	d := testNow.AddDate(0, 0, -daysAgo)
	return time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, time.Local)
}

var testFoods = []Food{
	{Name: "Oats", Kcal: 370, Protein: 13, Carbs: 60, Fat: 7, Fiber: 10, Sugar: 1, Salt: 0.01},
	{Name: "Banana", Kcal: 99, Protein: 1, Carbs: 23, Fat: 0.3, Fiber: 2.6, Sugar: 12},
}

func TestTotalsAdd(t *testing.T) {
	var total Totals
	total.Add(testFoods[0], 80) // 80 g of oats

	want := Totals{Kcal: 296, Protein: 10.4, Carbs: 48, Fat: 5.6, Fiber: 8, Sugar: 0.8, Salt: 0.008}
	if !approx(total.Kcal, want.Kcal) || !approx(total.Protein, want.Protein) ||
		!approx(total.Carbs, want.Carbs) || !approx(total.Fat, want.Fat) ||
		!approx(total.Fiber, want.Fiber) || !approx(total.Sugar, want.Sugar) ||
		!approx(total.Salt, want.Salt) {
		t.Errorf("got %+v, want %+v", total, want)
	}
}

func TestFindFood(t *testing.T) {
	f, ok := findFood(testFoods, "oats") // different upper/lower case
	if !ok || f.Name != "Oats" {
		t.Errorf("expected to find Oats, got %+v, %v", f, ok)
	}
	if _, ok := findFood(testFoods, "Pizza"); ok {
		t.Error("Pizza should not be found")
	}
}

func TestSumEntries(t *testing.T) {
	entries := []Log{
		{Food: "oats", Grams: 80},
		{Food: "Banana", Grams: 120},
		{Food: "Pizza", Grams: 50},
	}

	total, missing := sumEntries(testFoods, entries)

	if !approx(total.Kcal, 296+118.8) {
		t.Errorf("kcal: got %v, want %v", total.Kcal, 296+118.8)
	}
	if len(missing) != 1 || missing[0] != "Pizza" {
		t.Errorf("expected Pizza to be reported missing, got %v", missing)
	}
}

func TestPrintTodayShowsAllNutrients(t *testing.T) {
	entries := []Log{{Time: at(0), Food: "Oats", Grams: 80}}
	var out bytes.Buffer

	printToday(&out, testFoods, entries, testNow)

	got := out.String()
	for _, want := range []string{"Calories", "Protein", "Carbs", "sugar", "Fat", "Fiber", "Salt", "296"} {
		if !strings.Contains(got, want) {
			t.Errorf("today's output is missing %q:\n%s", want, got)
		}
	}
}

func TestPrintTodayUnknownFood(t *testing.T) {
	entries := []Log{{Time: at(0), Food: "Pizza", Grams: 200}}
	var out bytes.Buffer

	printToday(&out, testFoods, entries, testNow)

	if !strings.Contains(out.String(), "Not in food list") {
		t.Errorf("expected a warning about Pizza:\n%s", out.String())
	}
}

func TestPrintTodayEmpty(t *testing.T) {
	var out bytes.Buffer

	printToday(&out, testFoods, nil, testNow)

	if !strings.Contains(out.String(), "Nothing logged today.") {
		t.Errorf("expected the empty message:\n%s", out.String())
	}
}

func TestPrintWeek(t *testing.T) {
	entries := []Log{
		{Time: at(2), Food: "Oats", Grams: 100}, // 370 kcal, two days ago
		{Time: at(0), Food: "Oats", Grams: 80},  // today: must NOT be in the week table
	}
	var out bytes.Buffer

	printWeek(&out, testFoods, entries, testNow)

	got := out.String()
	for i := 1; i <= 7; i++ {
		date := testNow.AddDate(0, 0, -i).Format("Mon 02 Jan")
		if !strings.Contains(got, date) {
			t.Errorf("missing row for %s:\n%s", date, got)
		}
	}
	if !strings.Contains(got, "370") {
		t.Errorf("expected 370 kcal for two days ago:\n%s", got)
	}
	if strings.Contains(got, "296") {
		t.Errorf("today's entry should not appear in the week table:\n%s", got)
	}
	if !strings.Contains(got, "-") {
		t.Errorf("days with no entries should show '-':\n%s", got)
	}
}
