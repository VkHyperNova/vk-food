package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"
)

func Report() {

	ClearScreen()

	foods, err := LoadFood(FoodFile)
	if err != nil {
		fmt.Printf("LoadFood: %v\n", err)
	}

	logs, err := LoadLog(LogFile)
	if err != nil {
		fmt.Printf("LoadLog: %v\n", err)
	}

	PrintFoods()

	today := time.Now()
	printWeek(os.Stdout, foods, logs, today)

	fmt.Println()
	printToday(os.Stdout, foods, logs, today)
}

func PrintFoods() error {
	foods, err := LoadFood(FoodFile)
	if err != nil {
		return err
	}

	var buf strings.Builder
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tKcal\tProtein\tCarbs\tFat\tFiber\tSugar\tSalt")
	for _, f := range foods {
		fmt.Fprintf(w, "%s\t%.1f\t%.1f\t%.1f\t%.1f\t%.1f\t%.1f\t%.2f\n",
			f.Name, f.Kcal, f.Protein, f.Carbs, f.Fat, f.Fiber, f.Sugar, f.Salt)
	}
	if err := w.Flush(); err != nil {
		return err
	}

	header, rows, _ := strings.Cut(buf.String(), "\n")
	fmt.Println(Bold + header + Reset)
	fmt.Print(rows)
	fmt.Println()
	return nil
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
func sumEntries(foods []Food, entries []Log) Totals {
	var total Totals
	for _, e := range entries {
		f, _ := findFood(foods, e.Food)

		total.Add(f, e.Grams)
	}
	return total
}

func printWeek(out io.Writer, foods []Food, entries []Log, today time.Time) {
	fmt.Fprintln(out, Bold+"Last 7 days"+Reset)

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Date\tkcal\tProtein\tCarbs\tFat\t")
	for i := 7; i >= 0; i-- {
		day := today.AddDate(0, 0, -i)
		date := day.Format("Mon 02 Jan")

		list := LogsOn(entries, day)
		if len(list) == 0 {
			fmt.Fprintf(w, "%s\t-\t-\t-\t-\t\n", date)
			continue
		}
		t := sumEntries(foods, list)
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

	total := sumEntries(foods, list)
	fmt.Fprintln(out)

	kcal := levelColor(total.Kcal, 1000, 2400, 4000)
	fmt.Fprintf(out, "  Calories   %s%7.2f g%s / 4000 g\n", kcal, total.Kcal, Reset)

	protein := levelColor(total.Protein, 85, 150, 180)
	fmt.Fprintf(out, "  Protein    %s%7.2f g%s / 180 g\n", protein, total.Protein, Reset)

	carbs := levelColor(total.Carbs, 50, 150, 300)
	fmt.Fprintf(out, "  Carbs      %s%7.2f g%s / 300 g\n", carbs, total.Carbs, Reset)

	sugar := levelColor(total.Sugar, 0, 25, 50)
	fmt.Fprintf(out, "  Sugar      %s%7.2f g%s / 50 g\n", sugar, total.Sugar, Reset)

	fat := levelColor(total.Fat, 45, 75, 90)
	fmt.Fprintf(out, "  Fat        %s%7.2f g%s / 90 g\n", fat, total.Fat, Reset)

	fiber := levelColor(total.Fiber, 25, 41, 50)
	fmt.Fprintf(out, "  Fiber      %s%7.2f g%s / 50 g\n", fiber, total.Fiber, Reset)

	salt := levelColor(total.Salt, 1, 5, 10)
	fmt.Fprintf(out, "  Salt       %s%7.2f g%s / 10 g\n", salt, total.Salt, Reset)
}

func levelColor(v, greenFrom, yellowFrom, redFrom float64) string {
	switch {
	case v > redFrom:
		return Bold + Red
	case v > yellowFrom:
		return Bold + Yellow
	case v >= greenFrom:
		return Bold + Green
	default:
		return Reset
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

func ClearScreen() {

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error clearing screen:", err)
	}
}
