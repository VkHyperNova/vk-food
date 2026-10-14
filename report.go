package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

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
func sumEntries(foods []Food, entries []Entry) (Totals, []string) {
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

func printWeek(foods []Food, entries []Entry, today time.Time) {
	fmt.Println(Bold + "Last 7 days" + Reset)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Date\tkcal\tProtein\tCarbs\tFat\t")
	for i := 7; i >= 1; i-- {
		day := today.AddDate(0, 0, -i)
		date := day.Format("Mon 02 Jan")

		list := EntriesOn(entries, day)
		if len(list) == 0 {
			fmt.Fprintf(w, "%s\t-\t-\t-\t-\t\n", date)
			continue
		}
		t, _ := sumEntries(foods, list)
		fmt.Fprintf(w, "%s\t%.0f\t%.1f\t%.1f\t%.1f\t\n", date, t.Kcal, t.Protein, t.Carbs, t.Fat)
	}
	w.Flush()
}

func printToday(foods []Food, entries []Entry, today time.Time) {
	fmt.Println(Bold + "Today, " + today.Format("Mon 02 Jan 2006") + Reset)

	list := EntriesOn(entries, today)
	if len(list) == 0 {
		fmt.Println("Nothing logged today.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
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
	fmt.Println()
	fmt.Printf("  Calories  %7.0f kcal\n", total.Kcal)
	fmt.Printf("  Protein   %7.1f g\n", total.Protein)
	fmt.Printf("  Carbs     %7.1f g\n", total.Carbs)
	fmt.Printf("    sugar   %7.1f g\n", total.Sugar)
	fmt.Printf("  Fat       %7.1f g\n", total.Fat)
	fmt.Printf("  Fiber     %7.1f g\n", total.Fiber)
	fmt.Printf("  Salt      %7.2f g\n", total.Salt)

	if len(missing) > 0 {
		fmt.Println(Yellow + "Not in food list (skipped): " + strings.Join(missing, ", ") + Reset)
	}
}
