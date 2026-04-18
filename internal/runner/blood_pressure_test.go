package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
	runner "github.com/yazanabuashour/openhealth/internal/runner"
)

func TestRunBloodPressureTaskRecordAndList(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}
	pulse64 := 64

	result, err := runner.RunBloodPressureTask(ctx, config, runner.BloodPressureTaskRequest{
		Action: runner.BloodPressureTaskActionRecord,
		Readings: []runner.BloodPressureInput{
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: &pulse64},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		},
	})
	if err != nil {
		t.Fatalf("run record task: %v", err)
	}
	if result.Rejected {
		t.Fatalf("result rejected: %#v", result)
	}
	if got := bloodPressureWriteStatuses(result.Writes); got != "created,created" {
		t.Fatalf("write statuses = %q, want created,created", got)
	}
	assertBloodPressureEntries(t, result.Entries, []runner.BloodPressureEntry{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: &pulse64},
	})

	listed, err := runner.RunBloodPressureTask(ctx, config, runner.BloodPressureTaskRequest{
		Action:   runner.BloodPressureTaskActionList,
		ListMode: runner.BloodPressureListModeHistory,
	})
	if err != nil {
		t.Fatalf("run list task: %v", err)
	}
	assertBloodPressureEntries(t, listed.Entries, []runner.BloodPressureEntry{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: &pulse64},
	})
}

func TestRunBloodPressureTaskCorrectsExistingReading(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}
	pulse64 := 64

	_, err := runner.RunBloodPressureTask(ctx, config, runner.BloodPressureTaskRequest{
		Action: runner.BloodPressureTaskActionRecord,
		Readings: []runner.BloodPressureInput{
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: &pulse64},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		},
	})
	if err != nil {
		t.Fatalf("seed blood pressure: %v", err)
	}

	corrected, err := runner.RunBloodPressureTask(ctx, config, runner.BloodPressureTaskRequest{
		Action: runner.BloodPressureTaskActionCorrect,
		Readings: []runner.BloodPressureInput{
			{Date: "2026-03-29", Systolic: 121, Diastolic: 77},
		},
	})
	if err != nil {
		t.Fatalf("run correction task: %v", err)
	}
	if corrected.Rejected {
		t.Fatalf("correction rejected: %#v", corrected)
	}
	if got := bloodPressureWriteStatuses(corrected.Writes); got != "updated" {
		t.Fatalf("correction write statuses = %q, want updated", got)
	}
	assertBloodPressureEntries(t, corrected.Entries, []runner.BloodPressureEntry{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		{Date: "2026-03-29", Systolic: 121, Diastolic: 77},
		{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
	})
}

func TestRunBloodPressureTaskCorrectRejectsMissingOrAmbiguousReading(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	missing, err := runner.RunBloodPressureTask(ctx, config, runner.BloodPressureTaskRequest{
		Action: runner.BloodPressureTaskActionCorrect,
		Readings: []runner.BloodPressureInput{
			{Date: "2026-03-29", Systolic: 121, Diastolic: 77},
		},
	})
	if err != nil {
		t.Fatalf("missing correction task: %v", err)
	}
	if !missing.Rejected || missing.RejectionReason != "no existing blood pressure reading for 2026-03-29" {
		t.Fatalf("missing result = %#v", missing)
	}

	_, err = runner.RunBloodPressureTask(ctx, config, runner.BloodPressureTaskRequest{
		Action: runner.BloodPressureTaskActionRecord,
		Readings: []runner.BloodPressureInput{
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
			{Date: "2026-03-29", Systolic: 120, Diastolic: 76},
		},
	})
	if err != nil {
		t.Fatalf("seed ambiguous blood pressure: %v", err)
	}

	ambiguous, err := runner.RunBloodPressureTask(ctx, config, runner.BloodPressureTaskRequest{
		Action: runner.BloodPressureTaskActionCorrect,
		Readings: []runner.BloodPressureInput{
			{Date: "2026-03-29", Systolic: 121, Diastolic: 77},
		},
	})
	if err != nil {
		t.Fatalf("ambiguous correction task: %v", err)
	}
	if !ambiguous.Rejected || ambiguous.RejectionReason != "multiple blood pressure readings for 2026-03-29; correction is ambiguous" {
		t.Fatalf("ambiguous result = %#v", ambiguous)
	}
}

func TestRunBloodPressureTaskBoundedRangeLatestAndLimit(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	_, err := runner.RunBloodPressureTask(ctx, config, runner.BloodPressureTaskRequest{
		Action: runner.BloodPressureTaskActionRecord,
		Readings: []runner.BloodPressureInput{
			{Date: "2026-03-27", Systolic: 126, Diastolic: 82},
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		},
	})
	if err != nil {
		t.Fatalf("seed blood pressure: %v", err)
	}

	bounded, err := runner.RunBloodPressureTask(ctx, config, runner.BloodPressureTaskRequest{
		Action:   runner.BloodPressureTaskActionList,
		ListMode: runner.BloodPressureListModeRange,
		FromDate: "2026-03-29",
		ToDate:   "2026-03-30",
	})
	if err != nil {
		t.Fatalf("bounded range task: %v", err)
	}
	assertBloodPressureEntries(t, bounded.Entries, []runner.BloodPressureEntry{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
	})

	latest, err := runner.RunBloodPressureTask(ctx, config, runner.BloodPressureTaskRequest{
		Action:   runner.BloodPressureTaskActionList,
		ListMode: runner.BloodPressureListModeLatest,
	})
	if err != nil {
		t.Fatalf("latest task: %v", err)
	}
	assertBloodPressureEntries(t, latest.Entries, []runner.BloodPressureEntry{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
	})

	limited, err := runner.RunBloodPressureTask(ctx, config, runner.BloodPressureTaskRequest{
		Action:   runner.BloodPressureTaskActionList,
		ListMode: runner.BloodPressureListModeHistory,
		Limit:    2,
	})
	if err != nil {
		t.Fatalf("limited history task: %v", err)
	}
	assertBloodPressureEntries(t, limited.Entries, []runner.BloodPressureEntry{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
	})
}

func TestRunBloodPressureTaskRejectsInvalidInputBeforeOpeningDatabase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		reading runner.BloodPressureInput
		reason  string
	}{
		{
			name:    "short date",
			reading: runner.BloodPressureInput{Date: "03/29", Systolic: 122, Diastolic: 78},
			reason:  "date must be YYYY-MM-DD",
		},
		{
			name:    "missing date",
			reading: runner.BloodPressureInput{Systolic: 122, Diastolic: 78},
			reason:  "date must be YYYY-MM-DD",
		},
		{
			name:    "nonpositive systolic",
			reading: runner.BloodPressureInput{Date: "2026-03-29", Systolic: 0, Diastolic: 78},
			reason:  "systolic must be greater than 0",
		},
		{
			name:    "nonpositive diastolic",
			reading: runner.BloodPressureInput{Date: "2026-03-29", Systolic: 122, Diastolic: -1},
			reason:  "diastolic must be greater than 0",
		},
		{
			name:    "nonpositive pulse",
			reading: runner.BloodPressureInput{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(0)},
			reason:  "pulse must be greater than 0",
		},
		{
			name:    "duplicate correction date",
			reading: runner.BloodPressureInput{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
			reason:  "duplicate correction date 2026-03-29",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dbPath := filepath.Join(t.TempDir(), "openhealth.db")
			request := runner.BloodPressureTaskRequest{
				Action:   runner.BloodPressureTaskActionRecord,
				Readings: []runner.BloodPressureInput{tt.reading},
			}
			if tt.name == "duplicate correction date" {
				request.Action = runner.BloodPressureTaskActionCorrect
				request.Readings = append(request.Readings, tt.reading)
			}
			result, err := runner.RunBloodPressureTask(context.Background(), client.LocalConfig{DatabasePath: dbPath}, request)
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

func bloodPressureWriteStatuses(writes []runner.BloodPressureWrite) string {
	out := ""
	for i, write := range writes {
		if i > 0 {
			out += ","
		}
		out += write.Status
	}
	return out
}

func assertBloodPressureEntries(t *testing.T, got []runner.BloodPressureEntry, want []runner.BloodPressureEntry) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("entry count = %d (%#v), want %d (%#v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i].Date != want[i].Date ||
			got[i].Systolic != want[i].Systolic ||
			got[i].Diastolic != want[i].Diastolic ||
			!equalIntPointer(got[i].Pulse, want[i].Pulse) {
			t.Fatalf("entry %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func equalIntPointer(left *int, right *int) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func intPointer(value int) *int {
	return &value
}
