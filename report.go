package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

func Report(foods []Food, entries []Log) {
	today := time.Now()
	printWeek(os.Stdout, foods, entries, today)
	fmt.Println()
	printToday(os.Stdout, foods, entries, today)
}

// Totals is a sum of nutrients (absolute amounts, not per 100 g).
type Totals struct {
	Kcal, Protein, Carbs, Fat, Fiber, Sugar, Salt float64
}

// Add adds the nutrients in grams of food f to the totals.
func (t *Totals) Add(f Food, grams float64) {
	k := grams / 100
	t.Kcal += f.Kcal * k
	t.Protein += f.Protein * k
	t.Carbs += f.Carbs * k
	t.Fat += f.Fat * k
	t.Fiber += f.Fiber * k
	t.Sugar += f.Sugar * k
	t.Salt += f.Salt * k
}

// findFood looks up a food by name, ignoring upper/lower case.
func findFood(foods []Food, name string) (Food, bool) {
	for _, f := range foods {
		if strings.EqualFold(f.Name, name) {
			return f, true
		}
	}
	return Food{}, false
}

// sumEntries adds up the nutrients of entries. Entries whose food is not in
// the food list are skipped, and their names are returned.
func sumEntries(foods []Food, entries []Log) (Totals, []string) {
	var total Totals
	var missing []string
	for _, e := range entries {
		f, ok := findFood(foods, e.Food)
		if !ok {
			missing = append(missing, e.Food)
			continue
		}
		total.Add(f, e.Grams)
	}
	return total, missing
}

func printWeek(out io.Writer, foods []Food, entries []Log, today time.Time) {
	fmt.Fprintln(out, Bold+"Last 7 days"+Reset)

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Date\tkcal\tProtein\tCarbs\tFat\t")
	for i := 7; i >= 1; i-- {
		day := today.AddDate(0, 0, -i)
		date := day.Format("Mon 02 Jan")

		list := LogsOn(entries, day)
		if len(list) == 0 {
			fmt.Fprintf(w, "%s\t-\t-\t-\t-\t\n", date)
			continue
		}
		t, _ := sumEntries(foods, list)
		fmt.Fprintf(w, "%s\t%.0f\t%.1f\t%.1f\t%.1f\t\n", date, t.Kcal, t.Protein, t.Carbs, t.Fat)
	}
	w.Flush()
}

func printToday(out io.Writer, foods []Food, entries []Log, today time.Time) {
	fmt.Fprintln(out, Bold+"Today, "+today.Format("Mon 02 Jan 2006")+Reset)

	list := LogsOn(entries, today)
	if len(list) == 0 {
		fmt.Fprintln(out, "Nothing logged today.")
		return
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Time\tFood\tGrams\tkcal\t")
	for _, e := range list {
		clock := e.Time.Local().Format("15:04")
		f, ok := findFood(foods, e.Food)
		if !ok {
			fmt.Fprintf(w, "%s\t%s\t%.0f\t?\t\n", clock, e.Food, e.Grams)
			continue
		}
		var t Totals
		t.Add(f, e.Grams)
		fmt.Fprintf(w, "%s\t%s\t%.0f\t%.0f\t\n", clock, e.Food, e.Grams, t.Kcal)
	}
	w.Flush()

	total, missing := sumEntries(foods, list)
	fmt.Fprintln(out)
	fmt.Fprintf(out, "  Calories  %7.0f kcal\n", total.Kcal)
	fmt.Fprintf(out, "  Protein   %7.1f g\n", total.Protein)
	fmt.Fprintf(out, "  Carbs     %7.1f g\n", total.Carbs)
	fmt.Fprintf(out, "    sugar   %7.1f g\n", total.Sugar)
	fmt.Fprintf(out, "  Fat       %7.1f g\n", total.Fat)
	fmt.Fprintf(out, "  Fiber     %7.1f g\n", total.Fiber)
	fmt.Fprintf(out, "  Salt      %7.2f g\n", total.Salt)

	if len(missing) > 0 {
		fmt.Fprintln(out, Yellow+"Not in food list (skipped): "+strings.Join(missing, ", ")+Reset)
	}
}

// sameDay reports whether two times fall on the same calendar day (local time).
func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Local().Date()
	by, bm, bd := b.Local().Date()
	return ay == by && am == bm && ad == bd
}

// LogsOn returns the entries logged on the given day.
func LogsOn(logs []Log, day time.Time) []Log {
	var out []Log
	for _, e := range logs {
		if sameDay(e.Time, day) {
			out = append(out, e)
		}
	}
	return out
}

// TodayLogs returns only the entries logged today.
func TodayLogs(logs []Log) []Log {
	return LogsOn(logs, time.Now())
}
