package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	client "github.com/yazanabuashour/openhealth/client"
	runner "github.com/yazanabuashour/openhealth/internal/runner"
)

func TestRunSleepTaskUpsertListAndDelete(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	result, err := runner.RunSleepTask(ctx, config, runner.SleepTaskRequest{
		Action: runner.SleepTaskActionUpsert,
		Entries: []runner.SleepInput{
			{Date: "2026-03-29", QualityScore: 4, WakeupCount: intPointer(2), Note: stringPointer(" woke up after storm ")},
			{Date: "2026-03-30", QualityScore: 5},
		},
	})
	if err != nil {
		t.Fatalf("run upsert task: %v", err)
	}
	if result.Rejected {
		t.Fatalf("result rejected: %#v", result)
	}
	if got := sleepWriteStatuses(result.Writes); got != "created,created" {
		t.Fatalf("write statuses = %q, want created,created", got)
	}
	assertSleepEntries(t, result.Entries, []runner.SleepTaskEntry{
		{Date: "2026-03-30", QualityScore: 5},
		{Date: "2026-03-29", QualityScore: 4, WakeupCount: intPointer(2), Note: stringPointer("woke up after storm")},
	})

	again, err := runner.RunSleepTask(ctx, config, runner.SleepTaskRequest{
		Action: runner.SleepTaskActionUpsert,
		Entries: []runner.SleepInput{
			{Date: "2026-03-29", QualityScore: 4},
		},
	})
	if err != nil {
		t.Fatalf("run repeat upsert task: %v", err)
	}
	if got := sleepWriteStatuses(again.Writes); got != "already_exists" {
		t.Fatalf("repeat write statuses = %q, want already_exists", got)
	}

	corrected, err := runner.RunSleepTask(ctx, config, runner.SleepTaskRequest{
		Action: runner.SleepTaskActionUpsert,
		Entries: []runner.SleepInput{
			{Date: "2026-03-29", QualityScore: 3, WakeupCount: intPointer(0)},
		},
	})
	if err != nil {
		t.Fatalf("run correction task: %v", err)
	}
	if got := sleepWriteStatuses(corrected.Writes); got != "updated" {
		t.Fatalf("correction write statuses = %q, want updated", got)
	}

	bounded, err := runner.RunSleepTask(ctx, config, runner.SleepTaskRequest{
		Action:   runner.SleepTaskActionList,
		ListMode: runner.SleepListModeRange,
		FromDate: "2026-03-29",
		ToDate:   "2026-03-29",
	})
	if err != nil {
		t.Fatalf("bounded sleep list: %v", err)
	}
	assertSleepEntries(t, bounded.Entries, []runner.SleepTaskEntry{
		{Date: "2026-03-29", QualityScore: 3, WakeupCount: intPointer(0), Note: stringPointer("woke up after storm")},
	})

	deleted, err := runner.RunSleepTask(ctx, config, runner.SleepTaskRequest{
		Action: runner.SleepTaskActionDelete,
		Target: &runner.SleepTarget{Date: "2026-03-29"},
	})
	if err != nil {
		t.Fatalf("delete sleep: %v", err)
	}
	if got := sleepWriteStatuses(deleted.Writes); got != "deleted" {
		t.Fatalf("delete write statuses = %q, want deleted", got)
	}
	assertSleepEntries(t, deleted.Entries, []runner.SleepTaskEntry{
		{Date: "2026-03-30", QualityScore: 5},
	})
}

func TestRunSleepTaskRejectsInvalidInputBeforeOpeningDatabase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  runner.SleepInput
		reason string
	}{
		{name: "short date", input: runner.SleepInput{Date: "03/29", QualityScore: 4}, reason: "date must be YYYY-MM-DD"},
		{name: "low quality", input: runner.SleepInput{Date: "2026-03-29", QualityScore: 0}, reason: "quality_score must be between 1 and 5"},
		{name: "high quality", input: runner.SleepInput{Date: "2026-03-29", QualityScore: 6}, reason: "quality_score must be between 1 and 5"},
		{name: "negative wakeup count", input: runner.SleepInput{Date: "2026-03-29", QualityScore: 4, WakeupCount: intPointer(-1)}, reason: "wakeup_count must be greater than or equal to 0"},
		{name: "empty note", input: runner.SleepInput{Date: "2026-03-29", QualityScore: 4, Note: stringPointer(" ")}, reason: "note must not be empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dbPath := filepath.Join(t.TempDir(), "openhealth.db")
			result, err := runner.RunSleepTask(context.Background(), client.LocalConfig{DatabasePath: dbPath}, runner.SleepTaskRequest{
				Action:  runner.SleepTaskActionUpsert,
				Entries: []runner.SleepInput{tt.input},
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

func sleepWriteStatuses(writes []runner.SleepWrite) string {
	out := ""
	for i, write := range writes {
		if i > 0 {
			out += ","
		}
		out += write.Status
	}
	return out
}

func assertSleepEntries(t *testing.T, got []runner.SleepTaskEntry, want []runner.SleepTaskEntry) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("entry count = %d (%#v), want %d (%#v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i].Date != want[i].Date ||
			got[i].QualityScore != want[i].QualityScore ||
			!equalIntPointer(got[i].WakeupCount, want[i].WakeupCount) ||
			!equalStringPointers(got[i].Note, want[i].Note) {
			t.Fatalf("entry %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}
