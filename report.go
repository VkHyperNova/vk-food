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

/* Commands */

func Report() {

	clearScreen()

	foods, err := LoadFood(Foods)
	if err != nil {
		fmt.Printf("LoadFood: %v\n", err)
	}

	logs, err := LoadLog(Logs)
	if err != nil {
		fmt.Printf("LoadLog: %v\n", err)
	}

	today := time.Now()
	printWeek(os.Stdout, foods, logs, today)

	fmt.Println()
	printToday(os.Stdout, foods, logs, today)
}

func printWeek(out io.Writer, foods []Food, entries []Log, today time.Time) {
	fmt.Fprintln(out, Bold+"Last 7 days"+Reset)

	fmt.Fprintf(out, "%10s%9s%9s%9s%9s%9s%9s%9s\n", "Date", "kcal", "Protein", "Carbs", "Fat", "Fiber", "Sugar", "Salt")
	for i := 7; i >= 0; i-- {
		day := today.AddDate(0, 0, -i)
		date := day.Format("Mon 02 Jan")

		list := logsOn(entries, day)
		if len(list) == 0 {
			fmt.Fprintf(out, "%10s%s\n", date, strings.Repeat(fmt.Sprintf("%9s", "-"), 7))
			continue
		}
		t := sumEntries(foods, list)
		c := levelsOf(t)
		fmt.Fprintf(out, "%10s%s%s%s%s%s%s%s\n", date,
			cell(c.Kcal, fmt.Sprintf("%.0f", t.Kcal)),
			cell(c.Protein, fmt.Sprintf("%.1f", t.Protein)),
			cell(c.Carbs, fmt.Sprintf("%.1f", t.Carbs)),
			cell(c.Fat, fmt.Sprintf("%.1f", t.Fat)),
			cell(c.Fiber, fmt.Sprintf("%.1f", t.Fiber)),
			cell(c.Sugar, fmt.Sprintf("%.1f", t.Sugar)),
			cell(c.Salt, fmt.Sprintf("%.1f", t.Salt)))
	}
}

func printToday(out io.Writer, foods []Food, entries []Log, today time.Time) {
	fmt.Fprintln(out, Bold+"Today, "+today.Format("Mon 02 Jan 2006")+Reset)

	list := logsOn(entries, today)
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
		t.add(f, e.Grams)
		fmt.Fprintf(w, "%s\t%s\t%.0f\t%.0f\t\n", clock, e.Food, e.Grams, t.Kcal)
	}
	w.Flush()

	total := sumEntries(foods, list)
	fmt.Fprintln(out)

	c := levelsOf(total)
	fmt.Fprintf(out, "  Calories %s%7.0f kcal%s (3100) \n", c.Kcal, total.Kcal, Reset)
	fmt.Fprintf(out, "  Protein    %s%7.2f g%s (180)\n", c.Protein, total.Protein, Reset)
	fmt.Fprintf(out, "  Carbs      %s%7.2f g%s (300)\n", c.Carbs, total.Carbs, Reset)
	fmt.Fprintf(out, "  Fat        %s%7.2f g%s (90)\n", c.Fat, total.Fat, Reset)
	fmt.Fprintf(out, "  Fiber      %s%7.2f g%s (50)\n", c.Fiber, total.Fiber, Reset)
	fmt.Fprintf(out, "  Free Sugar %s%7.2f g%s (50)\n", c.Sugar, total.Sugar, Reset)
	fmt.Fprintf(out, "  Salt       %s%7.2f g%s (5)\n", c.Salt, total.Salt, Reset)
}

/* Helper functions */

type Totals struct {
	Kcal, Protein, Carbs, Fat, Fiber, Sugar, Salt float64
}

func (t *Totals) add(f Food, grams float64) {
	k := grams / 100
	t.Kcal += f.Kcal * k
	t.Protein += f.Protein * k
	t.Carbs += f.Carbs * k
	t.Fat += f.Fat * k
	t.Fiber += f.Fiber * k
	t.Sugar += f.Sugar * k
	t.Salt += f.Salt * k
}

func findFood(foods []Food, name string) (Food, bool) {
	for _, f := range foods {
		if strings.EqualFold(f.Name, name) {
			return f, true
		}
	}
	return Food{}, false
}

func sumEntries(foods []Food, entries []Log) Totals {
	var total Totals
	for _, e := range entries {
		f, _ := findFood(foods, e.Food)

		total.add(f, e.Grams)
	}
	return total
}

func cell(color, text string) string {
	return color + fmt.Sprintf("%9s", text) + Reset
}

type Levels struct{ Kcal, Protein, Carbs, Fat, Fiber, Sugar, Salt string }

func levelsOf(t Totals) Levels {
	return Levels{
		Kcal:    levelColor(t.Kcal, 0, 2400, 3100),
		Protein: levelColor(t.Protein, 85, 150, 180),
		Carbs:   levelColor(t.Carbs, 50, 150, 300),
		Fat:     levelColor(t.Fat, 45, 75, 90),
		Fiber:   levelColor(t.Fiber, 25, 41, 50),
		Sugar:   levelColor(t.Sugar, 0, 25, 50),
		Salt:    levelColor(t.Salt, 0, 4, 5),
	}
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

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Local().Date()
	by, bm, bd := b.Local().Date()
	return ay == by && am == bm && ad == bd
}

func logsOn(logs []Log, day time.Time) []Log {
	var out []Log
	for _, e := range logs {
		if sameDay(e.Time, day) {
			out = append(out, e)
		}
	}
	return out
}

func clearScreen() {

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
