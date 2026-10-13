package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

const usage = `foodtracker - log what you eat and see calories and nutrients

Usage:
  food add-food <name> [--kcal N] [--protein N] [--carbs N] [--fat N]
                       [--fiber N] [--sugar N] [--salt N]     (values per 100 g)
  food foods                       list saved foods
  food log <food> <grams>          log something you ate
  food today                       show today's entries and totals
  food undo                        remove the last logged entry

Example:
  food add-food oats --kcal 370 --protein 13 --carbs 60 --fat 7 --fiber 10
  food log oats 80
  food today
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		return
	}
	var err error
	switch os.Args[1] {
	case "add-food":
		err = cmdAddFood(os.Args[2:])
	case "foods":
		err = cmdFoods()
	case "log":
		err = cmdLog(os.Args[2:])
	case "today":
		err = cmdToday()
	case "undo":
		err = cmdUndo()
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func cmdAddFood(args []string) error {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return fmt.Errorf("usage: food add-food <name> --kcal N --protein N ...")
	}
	name := args[0]

	fs := flag.NewFlagSet("add-food", flag.ContinueOnError)
	f := Food{Name: name}
	fs.Float64Var(&f.Kcal, "kcal", 0, "calories per 100 g")
	fs.Float64Var(&f.Protein, "protein", 0, "protein (g) per 100 g")
	fs.Float64Var(&f.Carbs, "carbs", 0, "carbohydrates (g) per 100 g")
	fs.Float64Var(&f.Fat, "fat", 0, "fat (g) per 100 g")
	fs.Float64Var(&f.Fiber, "fiber", 0, "fiber (g) per 100 g")
	fs.Float64Var(&f.Sugar, "sugar", 0, "sugar (g) per 100 g")
	fs.Float64Var(&f.Salt, "salt", 0, "salt (g) per 100 g")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	// If no calories were given, estimate them from the macros (4/4/9 rule).
	estimated := false
	if f.Kcal == 0 && (f.Protein > 0 || f.Carbs > 0 || f.Fat > 0) {
		f.Kcal = f.Protein*4 + f.Carbs*4 + f.Fat*9
		estimated = true
	}

	d, err := load()
	if err != nil {
		return err
	}
	replaced := d.upsertFood(f)
	if err := d.save(); err != nil {
		return err
	}

	verb := "Added"
	if replaced {
		verb = "Updated"
	}
	fmt.Printf("%s %s: %.0f kcal, %.1f g protein, %.1f g carbs, %.1f g fat per 100 g\n",
		verb, f.Name, f.Kcal, f.Protein, f.Carbs, f.Fat)
	if estimated {
		fmt.Println("(kcal estimated from macros: 4 per g protein/carbs, 9 per g fat)")
	}
	return nil
}

func cmdFoods() error {
	d, err := load()
	if err != nil {
		return err
	}
	if len(d.Foods) == 0 {
		fmt.Println("No foods yet. Add one with: food add-food oats --kcal 370 --protein 13 --carbs 60 --fat 7")
		return nil
	}
	foods := append([]Food(nil), d.Foods...)
	sort.Slice(foods, func(i, j int) bool {
		return strings.ToLower(foods[i].Name) < strings.ToLower(foods[j].Name)
	})
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Food\tkcal\tProt\tCarbs\tFat\tFiber\tSugar\tSalt\t")
	for _, f := range foods {
		fmt.Fprintf(w, "%s\t%.0f\t%.1f\t%.1f\t%.1f\t%.1f\t%.1f\t%.2f\t\n",
			f.Name, f.Kcal, f.Protein, f.Carbs, f.Fat, f.Fiber, f.Sugar, f.Salt)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	fmt.Println("(values per 100 g)")
	return nil
}

func cmdLog(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: food log <food> <grams>")
	}
	grams, err := strconv.ParseFloat(strings.TrimSuffix(args[1], "g"), 64)
	if err != nil || grams <= 0 {
		return fmt.Errorf("%q is not a valid amount in grams", args[1])
	}
	d, err := load()
	if err != nil {
		return err
	}
	food, err := d.findFood(args[0])
	if err != nil {
		return err
	}
	d.Entries = append(d.Entries, Entry{Time: time.Now(), Food: food.Name, Grams: grams})
	if err := d.save(); err != nil {
		return err
	}
	var t Totals
	t.Add(food, grams)
	fmt.Printf("Logged %.0f g %s: %.0f kcal\n", grams, food.Name, t.Kcal)
	return nil
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Local().Date()
	by, bm, bd := b.Local().Date()
	return ay == by && am == bm && ad == bd
}

func cmdToday() error {
	d, err := load()
	if err != nil {
		return err
	}
	now := time.Now()
	var total Totals

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Time\tFood\tGrams\tkcal\tProt\tCarbs\tFat\t")
	count := 0
	for _, e := range d.Entries {
		if !sameDay(e.Time, now) {
			continue
		}
		food, err := d.findFood(e.Food)
		if err != nil {
			// Food was removed from the list; skip but don't crash.
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			continue
		}
		var t Totals
		t.Add(food, e.Grams)
		total.Add(food, e.Grams)
		fmt.Fprintf(w, "%s\t%s\t%.0f\t%.0f\t%.1f\t%.1f\t%.1f\t\n",
			e.Time.Local().Format("15:04"), food.Name, e.Grams, t.Kcal, t.Protein, t.Carbs, t.Fat)
		count++
	}
	if count == 0 {
		fmt.Println("Nothing logged today. Try: food log oats 80")
		return nil
	}
	if err := w.Flush(); err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("Totals for %s\n", now.Format("Mon 2 Jan 2006"))
	fmt.Printf("  Calories  %7.0f kcal\n", total.Kcal)
	fmt.Printf("  Protein   %7.1f g\n", total.Protein)
	fmt.Printf("  Carbs     %7.1f g   (of which sugar %.1f g)\n", total.Carbs, total.Sugar)
	fmt.Printf("  Fat       %7.1f g\n", total.Fat)
	fmt.Printf("  Fiber     %7.1f g\n", total.Fiber)
	fmt.Printf("  Salt      %7.2f g\n", total.Salt)
	return nil
}

func cmdUndo() error {
	d, err := load()
	if err != nil {
		return err
	}
	if len(d.Entries) == 0 {
		fmt.Println("Nothing to undo.")
		return nil
	}
	last := d.Entries[len(d.Entries)-1]
	d.Entries = d.Entries[:len(d.Entries)-1]
	if err := d.save(); err != nil {
		return err
	}
	fmt.Printf("Removed: %.0f g %s (%s)\n", last.Grams, last.Food, last.Time.Local().Format("15:04"))
	return nil
}
