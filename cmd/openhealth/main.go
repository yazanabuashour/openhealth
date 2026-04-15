package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

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

	databasePath := fs.String("db", envOrDefault("OPENHEALTH_DATABASE_PATH", "openhealth.db"), "SQLite database path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("migrate does not accept positional arguments")
	}

	db, err := sqlite.Open(*databasePath)
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
	databasePath := fs.String("db", envOrDefault("OPENHEALTH_DATABASE_PATH", "openhealth.db"), "SQLite database path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("serve does not accept positional arguments")
	}

	db, err := sqlite.Open(*databasePath)
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

func writeUsage(w io.Writer) error {
	_, err := fmt.Fprintf(
		w,
		"%s\n\nUsage:\n  openhealth migrate [-db path]\n  openhealth serve [-listen addr] [-db path]\n",
		app.Banner(),
	)
	return err
}

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
