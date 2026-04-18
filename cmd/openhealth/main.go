package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
	runner "github.com/yazanabuashour/openhealth/internal/runner"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		_ = writeUsage(stderr)
		return errors.New("missing OpenHealth runner domain")
	}

	switch args[0] {
	case "help", "-h", "--help":
		return writeUsage(stdout)
	case "weight":
		return runWeight(args[1:], stdin, stdout)
	case "blood-pressure":
		return runBloodPressure(args[1:], stdin, stdout)
	case "medications":
		return runMedications(args[1:], stdin, stdout)
	case "labs":
		return runLabs(args[1:], stdin, stdout)
	default:
		_ = writeUsage(stderr)
		return fmt.Errorf("unknown OpenHealth runner domain %q", args[0])
	}
}

func runWeight(args []string, stdin io.Reader, stdout io.Writer) error {
	config, err := parseLocalConfig("weight", args)
	if err != nil {
		return err
	}

	var request runner.WeightTaskRequest
	if err := decodeRequest(stdin, &request); err != nil {
		return err
	}

	result, err := runner.RunWeightTask(context.Background(), config, request)
	if err != nil {
		return err
	}
	return encodeResult(stdout, result)
}

func runBloodPressure(args []string, stdin io.Reader, stdout io.Writer) error {
	config, err := parseLocalConfig("blood-pressure", args)
	if err != nil {
		return err
	}

	var request runner.BloodPressureTaskRequest
	if err := decodeRequest(stdin, &request); err != nil {
		return err
	}

	result, err := runner.RunBloodPressureTask(context.Background(), config, request)
	if err != nil {
		return err
	}
	return encodeResult(stdout, result)
}

func runMedications(args []string, stdin io.Reader, stdout io.Writer) error {
	config, err := parseLocalConfig("medications", args)
	if err != nil {
		return err
	}

	var request runner.MedicationTaskRequest
	if err := decodeRequest(stdin, &request); err != nil {
		return err
	}

	result, err := runner.RunMedicationTask(context.Background(), config, request)
	if err != nil {
		return err
	}
	return encodeResult(stdout, result)
}

func runLabs(args []string, stdin io.Reader, stdout io.Writer) error {
	config, err := parseLocalConfig("labs", args)
	if err != nil {
		return err
	}

	var request runner.LabTaskRequest
	if err := decodeRequest(stdin, &request); err != nil {
		return err
	}

	result, err := runner.RunLabTask(context.Background(), config, request)
	if err != nil {
		return err
	}
	return encodeResult(stdout, result)
}

func parseLocalConfig(name string, args []string) (client.LocalConfig, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	databasePath := fs.String("db", "", "SQLite database path")
	if err := fs.Parse(args); err != nil {
		return client.LocalConfig{}, err
	}
	if fs.NArg() != 0 {
		return client.LocalConfig{}, fmt.Errorf("%s does not accept positional arguments", name)
	}

	return client.LocalConfig{DatabasePath: *databasePath}, nil
}

func decodeRequest[T any](stdin io.Reader, request *T) error {
	decoder := json.NewDecoder(stdin)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(request); err != nil {
		return fmt.Errorf("decode request JSON: %w", err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("decode request JSON: multiple JSON values are not supported")
		}
		return fmt.Errorf("decode request JSON: %w", err)
	}
	return nil
}

func encodeResult[T any](stdout io.Writer, result T) error {
	encoder := json.NewEncoder(stdout)
	return encoder.Encode(result)
}

func writeUsage(w io.Writer) error {
	_, err := fmt.Fprint(w, `Usage:
  openhealth weight [-db path] < request.json
  openhealth blood-pressure [-db path] < request.json
  openhealth medications [-db path] < request.json
  openhealth labs [-db path] < request.json
`)
	return err
}
