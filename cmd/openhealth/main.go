package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yazanabuashour/openhealth/client"
	"github.com/yazanabuashour/openhealth/internal/app"
	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/httpapi"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		return writeUsage(stdout)
	}

	switch args[0] {
	case "help", "-h", "--help":
		return writeUsage(stdout)
	case "migrate":
		return runMigrate(args[1:], stdout)
	case "serve":
		return runServe(args[1:], stdout)
	case "weight":
		return runWeight(args[1:], stdout)
	case "blood-pressure":
		return runBloodPressure(args[1:], stdout)
	default:
		if err := writeUsage(stderr); err != nil {
			return err
		}
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runMigrate(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	databasePath := fs.String("db", "", "SQLite database path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("migrate does not accept positional arguments")
	}
	if *databasePath == "" {
		resolvedPath, err := resolveDefaultDatabasePath()
		if err != nil {
			return err
		}
		*databasePath = resolvedPath
	}

	db, err := openDatabase(*databasePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	if err := sqlite.ApplyMigrations(context.Background(), db); err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "migrations applied to %s\n", *databasePath)
	return err
}

func runServe(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	addr := fs.String("listen", envOrDefault("OPENHEALTH_LISTEN_ADDR", ":8080"), "HTTP listen address")
	databasePath := fs.String("db", "", "SQLite database path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("serve does not accept positional arguments")
	}
	if *databasePath == "" {
		resolvedPath, err := resolveDefaultDatabasePath()
		if err != nil {
			return err
		}
		*databasePath = resolvedPath
	}

	db, err := openDatabase(*databasePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	if err := sqlite.EnsureCurrent(context.Background(), db); err != nil {
		return err
	}

	repo := sqlite.NewRepository(db)
	service := health.NewService(repo)
	server := &http.Server{
		Addr:              *addr,
		Handler:           httpapi.NewHandler(service),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if _, err := fmt.Fprintf(stdout, "%s\nserving %s using %s\n", app.Banner(), *addr, *databasePath); err != nil {
		return err
	}

	return server.ListenAndServe()
}

func runWeight(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return writeWeightUsage(stdout)
	}
	switch args[0] {
	case "help", "-h", "--help":
		return writeWeightUsage(stdout)
	case "add":
		return runWeightAdd(args[1:], stdout)
	case "list":
		return runWeightList(args[1:], stdout)
	default:
		return fmt.Errorf("unknown weight command %q", args[0])
	}
}

func runBloodPressure(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return writeBloodPressureUsage(stdout)
	}
	switch args[0] {
	case "help", "-h", "--help":
		return writeBloodPressureUsage(stdout)
	case "add":
		return runBloodPressureAdd(args[1:], stdout)
	case "list":
		return runBloodPressureList(args[1:], stdout)
	default:
		return fmt.Errorf("unknown blood-pressure command %q", args[0])
	}
}

func runWeightAdd(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("weight add", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	databasePath := fs.String("db", "", "SQLite database path")
	dateValue := fs.String("date", "", "Recorded date in YYYY-MM-DD form")
	value := fs.Float64("value", 0, "Weight value")
	unit := fs.String("unit", string(client.WeightUnitLb), "Weight unit")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("weight add does not accept positional arguments")
	}

	recordedAt, err := parseCLIDateOnly(*dateValue)
	if err != nil {
		return err
	}
	if *value <= 0 {
		return fmt.Errorf("value must be greater than 0")
	}
	if *unit != string(client.WeightUnitLb) {
		return fmt.Errorf("unit must be lb")
	}

	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: *databasePath})
	if err != nil {
		return err
	}
	defer func() {
		_ = api.Close()
	}()

	result, err := api.UpsertWeight(context.Background(), client.WeightRecordInput{
		RecordedAt: recordedAt,
		Value:      *value,
		Unit:       client.WeightUnit(*unit),
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "%s %.1f %s %s\n", result.Entry.RecordedAt.Format(time.DateOnly), result.Entry.Value, result.Entry.Unit, result.Status)
	return err
}

func runWeightList(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("weight list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	databasePath := fs.String("db", "", "SQLite database path")
	fromValue := fs.String("from", "", "Start date in YYYY-MM-DD form")
	toValue := fs.String("to", "", "End date in YYYY-MM-DD form")
	limit := fs.Int("limit", 0, "Maximum number of rows")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("weight list does not accept positional arguments")
	}
	if *limit < 0 {
		return fmt.Errorf("limit must be greater than 0")
	}

	var from *time.Time
	if *fromValue != "" {
		parsed, err := parseCLIDateOnly(*fromValue)
		if err != nil {
			return err
		}
		from = &parsed
	}
	var to *time.Time
	if *toValue != "" {
		parsed, err := parseCLIDateOnly(*toValue)
		if err != nil {
			return err
		}
		endOfDay := parsed.Add(24*time.Hour - time.Nanosecond)
		to = &endOfDay
	}

	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: *databasePath})
	if err != nil {
		return err
	}
	defer func() {
		_ = api.Close()
	}()

	weights, err := api.ListWeights(context.Background(), client.WeightListOptions{
		From:  from,
		To:    to,
		Limit: *limit,
	})
	if err != nil {
		return err
	}
	for _, weight := range weights {
		if _, err := fmt.Fprintf(stdout, "%s %.1f %s\n", weight.RecordedAt.Format(time.DateOnly), weight.Value, weight.Unit); err != nil {
			return err
		}
	}
	return nil
}

func runBloodPressureAdd(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("blood-pressure add", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	databasePath := fs.String("db", "", "SQLite database path")
	dateValue := fs.String("date", "", "Recorded date in YYYY-MM-DD form")
	systolic := fs.Int("systolic", 0, "Systolic blood pressure")
	diastolic := fs.Int("diastolic", 0, "Diastolic blood pressure")
	pulse := fs.Int("pulse", 0, "Pulse")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("blood-pressure add does not accept positional arguments")
	}

	recordedAt, err := parseCLIDateOnly(*dateValue)
	if err != nil {
		return err
	}
	if *systolic <= 0 {
		return fmt.Errorf("systolic must be greater than 0")
	}
	if *diastolic <= 0 {
		return fmt.Errorf("diastolic must be greater than 0")
	}
	var pulseValue *int
	if pulseProvided(args) {
		if *pulse <= 0 {
			return fmt.Errorf("pulse must be greater than 0")
		}
		pulseValue = pulse
	}

	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: *databasePath})
	if err != nil {
		return err
	}
	defer func() {
		_ = api.Close()
	}()

	result, err := api.RecordBloodPressure(context.Background(), client.BloodPressureRecordInput{
		RecordedAt: recordedAt,
		Systolic:   *systolic,
		Diastolic:  *diastolic,
		Pulse:      pulseValue,
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "%s %d/%d%s created\n", result.RecordedAt.Format(time.DateOnly), result.Systolic, result.Diastolic, formatPulse(result.Pulse))
	return err
}

func runBloodPressureList(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("blood-pressure list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	databasePath := fs.String("db", "", "SQLite database path")
	fromValue := fs.String("from", "", "Start date in YYYY-MM-DD form")
	toValue := fs.String("to", "", "End date in YYYY-MM-DD form")
	limit := fs.Int("limit", 0, "Maximum number of rows")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("blood-pressure list does not accept positional arguments")
	}
	if *limit < 0 {
		return fmt.Errorf("limit must be greater than 0")
	}

	var from *time.Time
	if *fromValue != "" {
		parsed, err := parseCLIDateOnly(*fromValue)
		if err != nil {
			return err
		}
		from = &parsed
	}
	var to *time.Time
	if *toValue != "" {
		parsed, err := parseCLIDateOnly(*toValue)
		if err != nil {
			return err
		}
		endOfDay := parsed.Add(24*time.Hour - time.Nanosecond)
		to = &endOfDay
	}

	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: *databasePath})
	if err != nil {
		return err
	}
	defer func() {
		_ = api.Close()
	}()

	readings, err := api.ListBloodPressure(context.Background(), client.BloodPressureListOptions{
		From:  from,
		To:    to,
		Limit: *limit,
	})
	if err != nil {
		return err
	}
	for _, reading := range readings {
		if _, err := fmt.Fprintf(stdout, "%s %d/%d%s\n", reading.RecordedAt.Format(time.DateOnly), reading.Systolic, reading.Diastolic, formatPulse(reading.Pulse)); err != nil {
			return err
		}
	}
	return nil
}

func writeUsage(w io.Writer) error {
	_, err := fmt.Fprintf(
		w,
		"%s\n\nUsage:\n  openhealth migrate [-db path]\n  openhealth serve [-listen addr] [-db path]\n  openhealth weight add --date YYYY-MM-DD --value N [--unit lb] [-db path]\n  openhealth weight list [-db path] [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--limit N]\n  openhealth blood-pressure add --date YYYY-MM-DD --systolic N --diastolic N [--pulse N] [-db path]\n  openhealth blood-pressure list [-db path] [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--limit N]\n",
		app.Banner(),
	)
	return err
}

func writeWeightUsage(w io.Writer) error {
	_, err := fmt.Fprintf(
		w,
		"%s\n\nUsage:\n  openhealth weight add --date YYYY-MM-DD --value N [--unit lb] [-db path]\n  openhealth weight list [-db path] [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--limit N]\n",
		app.Banner(),
	)
	return err
}

func writeBloodPressureUsage(w io.Writer) error {
	_, err := fmt.Fprintf(
		w,
		"%s\n\nUsage:\n  openhealth blood-pressure add --date YYYY-MM-DD --systolic N --diastolic N [--pulse N] [-db path]\n  openhealth blood-pressure list [-db path] [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--limit N]\n",
		app.Banner(),
	)
	return err
}

func pulseProvided(args []string) bool {
	for _, arg := range args {
		if arg == "--pulse" || arg == "-pulse" ||
			strings.HasPrefix(arg, "--pulse=") || strings.HasPrefix(arg, "-pulse=") {
			return true
		}
	}
	return false
}

func formatPulse(pulse *int) string {
	if pulse == nil {
		return ""
	}
	return fmt.Sprintf(" pulse %d", *pulse)
}

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func resolveDefaultDatabasePath() (string, error) {
	paths, err := client.ResolveLocalPaths(client.LocalConfig{})
	if err != nil {
		return "", err
	}
	return paths.DatabasePath, nil
}

func openDatabase(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return sqlite.Open(path)
}

func parseCLIDateOnly(value string) (time.Time, error) {
	if len(value) != len(time.DateOnly) || value[4] != '-' || value[7] != '-' {
		return time.Time{}, fmt.Errorf("date must be YYYY-MM-DD")
	}
	for i, ch := range value {
		if i == 4 || i == 7 {
			continue
		}
		if ch < '0' || ch > '9' {
			return time.Time{}, fmt.Errorf("date must be YYYY-MM-DD")
		}
	}
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("date must be YYYY-MM-DD")
	}
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), nil
}
