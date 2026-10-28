# vk-food

A small food tracker for the terminal, written in Go.

You keep a list of foods with their nutrition per 100 g, then log what you eat in grams. The app adds it all up and shows the last 7 days and a detailed view of today, with colours that tell you when you are getting close to your limits.

It tracks **calories, protein, carbs, fat, fiber, sugar and salt**.

Data is stored in plain JSON files. After every save the app can also copy them to a USB drive (labelled `VK DATA`), asking you first.

## Contents

- [Quick start](#quick-start)
- [Commands](#commands)
- [The report](#the-report)
- [Colours and limits](#colours-and-limits)
- [Where the data lives](#where-the-data-lives)
- [Backup to a USB drive](#backup-to-a-usb-drive)
- [Project layout](#project-layout)
- [Tests](#tests)

## Quick start

You need Go (developed with Go 1.26).

```bash
# add a food (nutrition values are per 100 g)
go run . add-food

# log 50 g of it
go run . log oats 50

# or pick the food from a list with the arrow keys and type the grams
go run . log

# show the report (also shown after every command)
go run .
```

`go run . help` prints all commands.

## Commands

| Command | What it does |
|---|---|
| `go run .` | Shows the report. |
| `go run . log` | Opens a picker: choose a food with the arrow keys (type to filter), then enter the grams. |
| `go run . log <food> <grams>` | Logs a food, for example `log oats 50`. The food must already exist. |
| `go run . undo-log` | Removes the last log entry. |
| `go run . add-food` | Adds a new food. Asks for the name and the values per 100 g. |
| `go run . update-food <food>` | Changes a food. Each question shows the current value in green, and Enter keeps it. |
| `go run . delete-food <food>` | Deletes a food. Refused if the food is used in the log. |
| `go run . undo-food` | Removes the last added food. |
| `go run . print-foods` | Shows all foods with all their nutrition values. |
| `go run . help` | Shows the usage text. |

Things worth knowing:

- Food names are matched **without regard to upper and lower case**: `oats`, `Oats` and `OATS` are the same food. The log stores the name as it is written in the food list.
- A name with spaces needs quotes on the command line: `go run . log "peanut butter" 30`.
- `add-food` refuses an empty name and a name that already exists.
- `update-food` can also **rename** a food. All old log entries are renamed with it, so your history stays correct.
- `delete-food` refuses to delete a food that has log entries, so the report never shows unknown foods. Remove those entries first.
- Grams must be a number greater than zero. Use a dot for decimals (`12.5`, not `12,5`).
- Errors are printed in red at the bottom, below the report.

## The report

The report is shown after every command. It has two parts.

**Last 7 days** is one row per day (the 7 days before today and today), with a total at the bottom:

```
Last 7 days
Date      kcal Protein   Carbs     Fat   Fiber   Sugar    Salt
Wed 07       -       -       -       -       -       -       -
Thu 08       -       -       -       -       -       -       -
Fri 09       -       -       -       -       -       -       -
Sat 10       -       -       -       -       -       -       -
Sun 11       -       -       -       -       -       -       -
Mon 12    1340    51.0   140.0    64.0    26.0    11.0     0.1
Tue 13     370    13.0    60.0     7.0    10.0     1.0     0.0
Wed 14     365    14.0    36.0    18.5     6.8     3.2     0.0
--------------------------------------------------------------
Total     2075    78.0   236.0    89.5    42.8    15.2     0.1
```

**Today** is every entry of today (time, food, grams, kcal), followed by the totals for the day next to their limits:

```
Today, Wed 14 Oct 2026
   Time           Food  Grams  kcal
  08:30           Oats     50   185
  12:15  Peanut butter     30   180

  Calories     365 kcal (3100)
  Protein      14.00 g (180)
  Carbs        36.00 g (300)
  Fat          18.50 g (90)
  Fiber         6.80 g (50)
  Free Sugar    3.20 g (50)
  Salt          0.01 g (5)
```

If a log entry belongs to a food that is no longer in the food list, its kcal shows as `?` and it adds nothing to the totals.

Messages from saving (`Local Saved!`, `Backup Saved!`, `Backup skipped`) and the output of `print-foods` are shown **below** the report, because the report clears the screen before it prints.

## Colours and limits

The numbers in the report change colour as a day gets closer to its limit. White means below the first limit, then green, yellow, red.

| Nutrient | Green from | Yellow over | Red over |
|---|---|---|---|
| Calories (kcal) | 0 | 2400 | 3100 |
| Protein (g) | 85 | 150 | 180 |
| Carbs (g) | 50 | 150 | 300 |
| Fat (g) | 45 | 75 | 90 |
| Fiber (g) | 25 | 41 | 50 |
| Sugar (g) | 0 | 25 | 50 |
| Salt (g) | 0 | 4 | 5 |

The coloured numbers are also bold. The same limits are used for each row of the 7-day table and for today. The total row is bold only, because the limits are per day.

The limits are in one place, `levelsOf` in `report.go`. The numbers in brackets in the Today view are written in `printToday`, so change them there too.

## Where the data lives

Two JSON files, relative to the folder you run the program from:

```
DEVELOPMENT/FOOD/foods.json
DEVELOPMENT/FOOD/logs.json
```

The folder names are constants at the top of `main.go` (`BaseLocal`, `BaseDB`, `FoodsFile`, `LogsFile`). The folders are created when the first value is saved.

`foods.json` is a list of foods, with nutrition per 100 g:

```json
[
  {
    "Name": "Oats",
    "Kcal": 370,
    "Protein": 13,
    "Carbs": 60,
    "Fat": 7,
    "Fiber": 10,
    "Sugar": 1,
    "Salt": 0.02
  }
]
```

`logs.json` is a list of entries. The nutrition of an entry is worked out from the food list (`grams / 100 * value`) when the report is shown, so changing a food with `update-food` also changes the past days:

```json
[
  { "time": "2026-10-14T08:30:00+03:00", "food": "Oats", "grams": 50 }
]
```

A missing file is not an error: the first run starts with an empty list. A file that is not valid JSON is never overwritten. The command stops with an error instead.

## Backup to a USB drive

Every save first writes the local file. After that, the app tries to copy the same file to a USB drive labelled **`VK DATA`**:

- If the drive is already mounted, the file is copied.
- If it is not mounted, you are asked **Yes/No** whether to mount it. Yes mounts it and copies. No skips the backup (`Backup skipped`) and the local save stays.
- If mounting fails (for example the drive is unplugged), the error is shown. The local file is already saved.

The backup uses the same folders on the drive (`/media/veikko/VK DATA/DEVELOPMENT/FOOD/...`). The label and the mount point are the constants `driveLabel` and `driveMountPoint` in `main.go`.

Mounting is done by a separate package, [`github.com/VkHyperNova/vkmount`](https://github.com/VkHyperNova/vkmount), which finds the drive by its label and mounts it with `udisksctl`. It works on **Linux only** and needs `lsblk` and `udisksctl`. The app itself does not need the drive.

## Project layout

```
main.go            constants, file paths, usage text, main()
food.go            Food type, add / update / delete / undo / print foods, questions
log.go             Log type, log (command line and picker), undo-log
report.go          the report: 7-day table, today, limits and colours
storage.go         Save (local file, then backup) and writeJSON
food_test.go       tests for food.go
log_test.go        tests for log.go (and the shared test helpers)
report_test.go     tests for report.go
storage_test.go    tests for storage.go
```

Dependencies:

- [`charmbracelet/huh`](https://github.com/charmbracelet/huh) for the food picker and the Yes/No question
- [`vkmount`](https://github.com/VkHyperNova/vkmount) for mounting the backup drive

## Tests

```bash
go test .
```

Passing tests print nothing but `ok`. Failing ones are listed with a message. Add `-v` to see every test.

```bash
go test -v -run TestUpdateFood .                          # one group
go test -v -run 'TestUpdateFood/rename_changes_the_log' . # one test (spaces become _)
```

The tests never touch your real data and never ask a real question:

- the data files go to a temporary folder (`useTempData`)
- mounting the backup drive is replaced by a fake (`fakeMount`)
- standard input is replaced with a prepared text (`withStdin`), so the questions can be answered in a test
- `os.Args` is set for each test (`withArgs`), so commands run as if typed

Each test file is grouped by function, with a short comment at the top listing the groups, and the helpers at the bottom.

What is **not** tested, because it needs a terminal or a real drive: the arrow-key picker (`go run . log`), the real Yes/No prompt, `clearScreen`, and `main`. The mounting code has its own tests in the `vkmount` repository.