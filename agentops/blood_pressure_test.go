package agentops_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yazanabuashour/openhealth/agentops"
	"github.com/yazanabuashour/openhealth/client"
)

func TestRunBloodPressureTaskRecordAndList(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}
	pulse64 := 64

	result, err := agentops.RunBloodPressureTask(ctx, config, agentops.BloodPressureTaskRequest{
		Action: agentops.BloodPressureTaskActionRecord,
		Readings: []agentops.BloodPressureInput{
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
	assertBloodPressureEntries(t, result.Entries, []agentops.BloodPressureEntry{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: &pulse64},
	})

	listed, err := agentops.RunBloodPressureTask(ctx, config, agentops.BloodPressureTaskRequest{
		Action:   agentops.BloodPressureTaskActionList,
		ListMode: agentops.BloodPressureListModeHistory,
	})
	if err != nil {
		t.Fatalf("run list task: %v", err)
	}
	assertBloodPressureEntries(t, listed.Entries, []agentops.BloodPressureEntry{
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

	_, err := agentops.RunBloodPressureTask(ctx, config, agentops.BloodPressureTaskRequest{
		Action: agentops.BloodPressureTaskActionRecord,
		Readings: []agentops.BloodPressureInput{
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: &pulse64},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		},
	})
	if err != nil {
		t.Fatalf("seed blood pressure: %v", err)
	}

	corrected, err := agentops.RunBloodPressureTask(ctx, config, agentops.BloodPressureTaskRequest{
		Action: agentops.BloodPressureTaskActionCorrect,
		Readings: []agentops.BloodPressureInput{
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
	assertBloodPressureEntries(t, corrected.Entries, []agentops.BloodPressureEntry{
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

	missing, err := agentops.RunBloodPressureTask(ctx, config, agentops.BloodPressureTaskRequest{
		Action: agentops.BloodPressureTaskActionCorrect,
		Readings: []agentops.BloodPressureInput{
			{Date: "2026-03-29", Systolic: 121, Diastolic: 77},
		},
	})
	if err != nil {
		t.Fatalf("missing correction task: %v", err)
	}
	if !missing.Rejected || missing.RejectionReason != "no existing blood pressure reading for 2026-03-29" {
		t.Fatalf("missing result = %#v", missing)
	}

	_, err = agentops.RunBloodPressureTask(ctx, config, agentops.BloodPressureTaskRequest{
		Action: agentops.BloodPressureTaskActionRecord,
		Readings: []agentops.BloodPressureInput{
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
			{Date: "2026-03-29", Systolic: 120, Diastolic: 76},
		},
	})
	if err != nil {
		t.Fatalf("seed ambiguous blood pressure: %v", err)
	}

	ambiguous, err := agentops.RunBloodPressureTask(ctx, config, agentops.BloodPressureTaskRequest{
		Action: agentops.BloodPressureTaskActionCorrect,
		Readings: []agentops.BloodPressureInput{
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

	_, err := agentops.RunBloodPressureTask(ctx, config, agentops.BloodPressureTaskRequest{
		Action: agentops.BloodPressureTaskActionRecord,
		Readings: []agentops.BloodPressureInput{
			{Date: "2026-03-27", Systolic: 126, Diastolic: 82},
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		},
	})
	if err != nil {
		t.Fatalf("seed blood pressure: %v", err)
	}

	bounded, err := agentops.RunBloodPressureTask(ctx, config, agentops.BloodPressureTaskRequest{
		Action:   agentops.BloodPressureTaskActionList,
		ListMode: agentops.BloodPressureListModeRange,
		FromDate: "2026-03-29",
		ToDate:   "2026-03-30",
	})
	if err != nil {
		t.Fatalf("bounded range task: %v", err)
	}
	assertBloodPressureEntries(t, bounded.Entries, []agentops.BloodPressureEntry{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
	})

	latest, err := agentops.RunBloodPressureTask(ctx, config, agentops.BloodPressureTaskRequest{
		Action:   agentops.BloodPressureTaskActionList,
		ListMode: agentops.BloodPressureListModeLatest,
	})
	if err != nil {
		t.Fatalf("latest task: %v", err)
	}
	assertBloodPressureEntries(t, latest.Entries, []agentops.BloodPressureEntry{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
	})

	limited, err := agentops.RunBloodPressureTask(ctx, config, agentops.BloodPressureTaskRequest{
		Action:   agentops.BloodPressureTaskActionList,
		ListMode: agentops.BloodPressureListModeHistory,
		Limit:    2,
	})
	if err != nil {
		t.Fatalf("limited history task: %v", err)
	}
	assertBloodPressureEntries(t, limited.Entries, []agentops.BloodPressureEntry{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
	})
}

func TestRunBloodPressureTaskRejectsInvalidInputBeforeOpeningDatabase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		reading agentops.BloodPressureInput
		reason  string
	}{
		{
			name:    "short date",
			reading: agentops.BloodPressureInput{Date: "03/29", Systolic: 122, Diastolic: 78},
			reason:  "date must be YYYY-MM-DD",
		},
		{
			name:    "missing date",
			reading: agentops.BloodPressureInput{Systolic: 122, Diastolic: 78},
			reason:  "date must be YYYY-MM-DD",
		},
		{
			name:    "nonpositive systolic",
			reading: agentops.BloodPressureInput{Date: "2026-03-29", Systolic: 0, Diastolic: 78},
			reason:  "systolic must be greater than 0",
		},
		{
			name:    "nonpositive diastolic",
			reading: agentops.BloodPressureInput{Date: "2026-03-29", Systolic: 122, Diastolic: -1},
			reason:  "diastolic must be greater than 0",
		},
		{
			name:    "nonpositive pulse",
			reading: agentops.BloodPressureInput{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(0)},
			reason:  "pulse must be greater than 0",
		},
		{
			name:    "duplicate correction date",
			reading: agentops.BloodPressureInput{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
			reason:  "duplicate correction date 2026-03-29",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dbPath := filepath.Join(t.TempDir(), "openhealth.db")
			request := agentops.BloodPressureTaskRequest{
				Action:   agentops.BloodPressureTaskActionRecord,
				Readings: []agentops.BloodPressureInput{tt.reading},
			}
			if tt.name == "duplicate correction date" {
				request.Action = agentops.BloodPressureTaskActionCorrect
				request.Readings = append(request.Readings, tt.reading)
			}
			result, err := agentops.RunBloodPressureTask(context.Background(), client.LocalConfig{DatabasePath: dbPath}, request)
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

func bloodPressureWriteStatuses(writes []agentops.BloodPressureWrite) string {
	out := ""
	for i, write := range writes {
		if i > 0 {
			out += ","
		}
		out += write.Status
	}
	return out
}

func assertBloodPressureEntries(t *testing.T, got []agentops.BloodPressureEntry, want []agentops.BloodPressureEntry) {
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
