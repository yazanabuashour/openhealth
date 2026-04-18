package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
	runner "github.com/yazanabuashour/openhealth/internal/runner"
)

func TestRunWeightTaskUpsertAndList(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	result, err := runner.RunWeightTask(ctx, config, runner.WeightTaskRequest{
		Action: runner.WeightTaskActionUpsert,
		Weights: []runner.WeightInput{
			{Date: "2026-03-29", Value: 152.2, Unit: "lbs"},
			{Date: "2026-03-30", Value: 151.6, Unit: "pounds"},
		},
	})
	if err != nil {
		t.Fatalf("run upsert task: %v", err)
	}
	if result.Rejected {
		t.Fatalf("result rejected: %#v", result)
	}
	if got := writeStatuses(result.Writes); got != "created,created" {
		t.Fatalf("write statuses = %q, want created,created", got)
	}
	assertEntries(t, result.Entries, []runner.WeightTaskEntry{
		{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
	})

	again, err := runner.RunWeightTask(ctx, config, runner.WeightTaskRequest{
		Action: runner.WeightTaskActionUpsert,
		Weights: []runner.WeightInput{
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		},
	})
	if err != nil {
		t.Fatalf("run repeat upsert task: %v", err)
	}
	if got := writeStatuses(again.Writes); got != "already_exists,already_exists" {
		t.Fatalf("repeat write statuses = %q, want already_exists,already_exists", got)
	}

	corrected, err := runner.RunWeightTask(ctx, config, runner.WeightTaskRequest{
		Action: runner.WeightTaskActionUpsert,
		Weights: []runner.WeightInput{
			{Date: "2026-03-29", Value: 151.6, Unit: "lb"},
		},
	})
	if err != nil {
		t.Fatalf("run correction task: %v", err)
	}
	if got := writeStatuses(corrected.Writes); got != "updated" {
		t.Fatalf("correction write statuses = %q, want updated", got)
	}

	listed, err := runner.RunWeightTask(ctx, config, runner.WeightTaskRequest{
		Action:   runner.WeightTaskActionList,
		ListMode: runner.WeightListModeHistory,
	})
	if err != nil {
		t.Fatalf("run list task: %v", err)
	}
	assertEntries(t, listed.Entries, []runner.WeightTaskEntry{
		{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		{Date: "2026-03-29", Value: 151.6, Unit: "lb"},
	})
}

func TestRunWeightTaskBoundedRangeAndLatest(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	_, err := runner.RunWeightTask(ctx, config, runner.WeightTaskRequest{
		Action: runner.WeightTaskActionUpsert,
		Weights: []runner.WeightInput{
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		},
	})
	if err != nil {
		t.Fatalf("seed weights: %v", err)
	}

	bounded, err := runner.RunWeightTask(ctx, config, runner.WeightTaskRequest{
		Action:   runner.WeightTaskActionList,
		ListMode: runner.WeightListModeRange,
		FromDate: "2026-03-29",
		ToDate:   "2026-03-30",
	})
	if err != nil {
		t.Fatalf("bounded range task: %v", err)
	}
	assertEntries(t, bounded.Entries, []runner.WeightTaskEntry{
		{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
	})

	latest, err := runner.RunWeightTask(ctx, config, runner.WeightTaskRequest{
		Action:   runner.WeightTaskActionList,
		ListMode: runner.WeightListModeLatest,
	})
	if err != nil {
		t.Fatalf("latest task: %v", err)
	}
	assertEntries(t, latest.Entries, []runner.WeightTaskEntry{
		{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
	})
}

func TestRunWeightTaskRejectsInvalidInputBeforeOpeningDatabase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  runner.WeightInput
		reason string
	}{
		{
			name:   "short date",
			input:  runner.WeightInput{Date: "03/29", Value: 152.2, Unit: "lb"},
			reason: "date must be YYYY-MM-DD",
		},
		{
			name:   "nonpositive value",
			input:  runner.WeightInput{Date: "2026-03-29", Value: -5, Unit: "lb"},
			reason: "value must be greater than 0",
		},
		{
			name:   "unsupported unit",
			input:  runner.WeightInput{Date: "2026-03-29", Value: 152.2, Unit: "stone"},
			reason: "unit must be lb",
		},
		{
			name:   "missing unit",
			input:  runner.WeightInput{Date: "2026-03-29", Value: 152.2},
			reason: "unit must be lb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dbPath := filepath.Join(t.TempDir(), "openhealth.db")
			result, err := runner.RunWeightTask(context.Background(), client.LocalConfig{DatabasePath: dbPath}, runner.WeightTaskRequest{
				Action:  runner.WeightTaskActionUpsert,
				Weights: []runner.WeightInput{tt.input},
			})
			if err != nil {
				t.Fatalf("run task: %v", err)
			}
			if !result.Rejected || result.RejectionReason != tt.reason {
				t.Fatalf("result = %#v, want rejected reason %q", result, tt.reason)
			}
			if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
				t.Fatalf("database stat error = %v, want not exist", err)
			}
		})
	}
}

func writeStatuses(writes []runner.WeightWrite) string {
	out := ""
	for i, write := range writes {
		if i > 0 {
			out += ","
		}
		out += write.Status
	}
	return out
}

func assertEntries(t *testing.T, got []runner.WeightTaskEntry, want []runner.WeightTaskEntry) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("entry count = %d (%#v), want %d (%#v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("entry %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}
