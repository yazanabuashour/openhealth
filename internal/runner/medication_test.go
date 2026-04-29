package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	client "github.com/yazanabuashour/openhealth/client"
	runner "github.com/yazanabuashour/openhealth/internal/runner"
)

func TestRunMedicationTaskRecordListCorrectAndDelete(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	result, err := runner.RunMedicationTask(ctx, config, runner.MedicationTaskRequest{
		Action: runner.MedicationTaskActionRecord,
		Medications: []runner.MedicationInput{
			{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01", Note: stringPointer("Started after annual exam")},
			{Name: "Vitamin D", StartDate: "2026-02-01", EndDate: stringPointer("2026-03-01")},
		},
	})
	if err != nil {
		t.Fatalf("run record task: %v", err)
	}
	if result.Rejected {
		t.Fatalf("result rejected: %#v", result)
	}
	if got := medicationWriteStatuses(result.Writes); got != "created,created" {
		t.Fatalf("write statuses = %q, want created,created", got)
	}
	assertMedicationEntries(t, result.Entries, []runner.MedicationEntry{
		{Name: "Vitamin D", StartDate: "2026-02-01", EndDate: stringPointer("2026-03-01")},
		{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01", Note: stringPointer("Started after annual exam")},
	})

	again, err := runner.RunMedicationTask(ctx, config, runner.MedicationTaskRequest{
		Action: runner.MedicationTaskActionRecord,
		Medications: []runner.MedicationInput{
			{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01", Note: stringPointer("Started after annual exam")},
		},
	})
	if err != nil {
		t.Fatalf("repeat medication task: %v", err)
	}
	if got := medicationWriteStatuses(again.Writes); got != "already_exists" {
		t.Fatalf("repeat write statuses = %q, want already_exists", got)
	}

	conflict, err := runner.RunMedicationTask(ctx, config, runner.MedicationTaskRequest{
		Action: runner.MedicationTaskActionRecord,
		Medications: []runner.MedicationInput{
			{Name: "Levothyroxine", DosageText: stringPointer("50 mcg"), StartDate: "2026-01-01"},
		},
	})
	if err != nil {
		t.Fatalf("conflict medication task: %v", err)
	}
	if !conflict.Rejected || conflict.RejectionReason != "medication already exists for Levothyroxine starting 2026-01-01; use correct_medication" {
		t.Fatalf("conflict result = %#v", conflict)
	}

	corrected, err := runner.RunMedicationTask(ctx, config, runner.MedicationTaskRequest{
		Action: runner.MedicationTaskActionCorrect,
		Target: &runner.MedicationTarget{Name: "Levothyroxine", StartDate: "2026-01-01"},
		Medication: &runner.MedicationInput{
			Name:       "Levothyroxine",
			DosageText: stringPointer("50 mcg"),
			StartDate:  "2026-01-01",
			Note:       stringPointer("Dose increased after follow-up"),
		},
	})
	if err != nil {
		t.Fatalf("correct medication task: %v", err)
	}
	if corrected.Rejected {
		t.Fatalf("correction rejected: %#v", corrected)
	}
	if got := medicationWriteStatuses(corrected.Writes); got != "updated" {
		t.Fatalf("correction write statuses = %q, want updated", got)
	}
	assertMedicationEntries(t, corrected.Entries, []runner.MedicationEntry{
		{Name: "Vitamin D", StartDate: "2026-02-01", EndDate: stringPointer("2026-03-01")},
		{Name: "Levothyroxine", DosageText: stringPointer("50 mcg"), StartDate: "2026-01-01", Note: stringPointer("Dose increased after follow-up")},
	})

	correctedNoNote, err := runner.RunMedicationTask(ctx, config, runner.MedicationTaskRequest{
		Action: runner.MedicationTaskActionCorrect,
		Target: &runner.MedicationTarget{Name: "Levothyroxine", StartDate: "2026-01-01"},
		Medication: &runner.MedicationInput{
			Name:       "Levothyroxine",
			DosageText: stringPointer("75 mcg"),
			StartDate:  "2026-01-01",
		},
	})
	if err != nil {
		t.Fatalf("correct medication without note: %v", err)
	}
	if got := medicationWriteStatuses(correctedNoNote.Writes); got != "updated" {
		t.Fatalf("no-note correction write statuses = %q, want updated", got)
	}
	assertMedicationEntries(t, correctedNoNote.Entries, []runner.MedicationEntry{
		{Name: "Vitamin D", StartDate: "2026-02-01", EndDate: stringPointer("2026-03-01")},
		{Name: "Levothyroxine", DosageText: stringPointer("75 mcg"), StartDate: "2026-01-01", Note: stringPointer("Dose increased after follow-up")},
	})

	deleted, err := runner.RunMedicationTask(ctx, config, runner.MedicationTaskRequest{
		Action: runner.MedicationTaskActionDelete,
		Target: &runner.MedicationTarget{Name: "Vitamin D", StartDate: "2026-02-01"},
	})
	if err != nil {
		t.Fatalf("delete medication task: %v", err)
	}
	if deleted.Rejected {
		t.Fatalf("deletion rejected: %#v", deleted)
	}
	if got := medicationWriteStatuses(deleted.Writes); got != "deleted" {
		t.Fatalf("delete write statuses = %q, want deleted", got)
	}
	assertMedicationEntries(t, deleted.Entries, []runner.MedicationEntry{
		{Name: "Levothyroxine", DosageText: stringPointer("75 mcg"), StartDate: "2026-01-01", Note: stringPointer("Dose increased after follow-up")},
	})

	active, err := runner.RunMedicationTask(ctx, config, runner.MedicationTaskRequest{
		Action: runner.MedicationTaskActionList,
		Status: runner.MedicationStatusActive,
	})
	if err != nil {
		t.Fatalf("list active medications: %v", err)
	}
	assertMedicationEntries(t, active.Entries, []runner.MedicationEntry{
		{Name: "Levothyroxine", DosageText: stringPointer("75 mcg"), StartDate: "2026-01-01", Note: stringPointer("Dose increased after follow-up")},
	})
}

func TestRunMedicationTaskRejectsMissingOrAmbiguousTarget(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	missing, err := runner.RunMedicationTask(ctx, config, runner.MedicationTaskRequest{
		Action: runner.MedicationTaskActionCorrect,
		Target: &runner.MedicationTarget{Name: "Levothyroxine", StartDate: "2026-01-01"},
		Medication: &runner.MedicationInput{
			Name:      "Levothyroxine",
			StartDate: "2026-01-01",
		},
	})
	if err != nil {
		t.Fatalf("missing correction: %v", err)
	}
	if !missing.Rejected || missing.RejectionReason != "no matching medication" {
		t.Fatalf("missing result = %#v", missing)
	}

	api, err := client.OpenLocal(config)
	if err != nil {
		t.Fatalf("open local client: %v", err)
	}
	_, err = api.CreateMedicationCourse(ctx, client.MedicationCourseInput{Name: "Levothyroxine", StartDate: "2026-01-01"})
	if err != nil {
		t.Fatalf("create first medication: %v", err)
	}
	_, err = api.CreateMedicationCourse(ctx, client.MedicationCourseInput{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01"})
	if err != nil {
		t.Fatalf("create second medication: %v", err)
	}
	if err := api.Close(); err != nil {
		t.Fatalf("close local client: %v", err)
	}

	ambiguous, err := runner.RunMedicationTask(ctx, config, runner.MedicationTaskRequest{
		Action: runner.MedicationTaskActionDelete,
		Target: &runner.MedicationTarget{Name: "Levothyroxine", StartDate: "2026-01-01"},
	})
	if err != nil {
		t.Fatalf("ambiguous delete: %v", err)
	}
	if !ambiguous.Rejected || ambiguous.RejectionReason != "multiple matching medications; target is ambiguous" {
		t.Fatalf("ambiguous result = %#v", ambiguous)
	}
}

func TestRunMedicationTaskRejectsInvalidInputBeforeOpeningDatabase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		request runner.MedicationTaskRequest
		reason  string
	}{
		{
			name: "short start date",
			request: runner.MedicationTaskRequest{
				Action:      runner.MedicationTaskActionRecord,
				Medications: []runner.MedicationInput{{Name: "Levothyroxine", StartDate: "03/29"}},
			},
			reason: "start_date must be YYYY-MM-DD",
		},
		{
			name: "empty dosage",
			request: runner.MedicationTaskRequest{
				Action:      runner.MedicationTaskActionRecord,
				Medications: []runner.MedicationInput{{Name: "Levothyroxine", DosageText: stringPointer(" "), StartDate: "2026-01-01"}},
			},
			reason: "dosage_text must not be empty",
		},
		{
			name: "empty note",
			request: runner.MedicationTaskRequest{
				Action:      runner.MedicationTaskActionRecord,
				Medications: []runner.MedicationInput{{Name: "Levothyroxine", StartDate: "2026-01-01", Note: stringPointer(" ")}},
			},
			reason: "note must not be empty",
		},
		{
			name: "end before start",
			request: runner.MedicationTaskRequest{
				Action:      runner.MedicationTaskActionRecord,
				Medications: []runner.MedicationInput{{Name: "Levothyroxine", StartDate: "2026-01-02", EndDate: stringPointer("2026-01-01")}},
			},
			reason: "end_date must be on or after start_date",
		},
		{
			name: "unsupported status",
			request: runner.MedicationTaskRequest{
				Action: runner.MedicationTaskActionList,
				Status: "completed",
			},
			reason: "status must be active or all",
		},
		{
			name: "missing target",
			request: runner.MedicationTaskRequest{
				Action:     runner.MedicationTaskActionDelete,
				Target:     &runner.MedicationTarget{Name: "Levothyroxine"},
				Medication: &runner.MedicationInput{Name: "Levothyroxine", StartDate: "2026-01-01"},
			},
			reason: "target id or name and start_date are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dbPath := filepath.Join(t.TempDir(), "openhealth.db")
			result, err := runner.RunMedicationTask(context.Background(), client.LocalConfig{DatabasePath: dbPath}, tt.request)
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

func medicationWriteStatuses(writes []runner.MedicationWrite) string {
	out := ""
	for i, write := range writes {
		if i > 0 {
			out += ","
		}
		out += write.Status
	}
	return out
}

func assertMedicationEntries(t *testing.T, got []runner.MedicationEntry, want []runner.MedicationEntry) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("entry count = %d (%#v), want %d (%#v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i].Name != want[i].Name ||
			got[i].StartDate != want[i].StartDate ||
			!equalStringPointers(got[i].DosageText, want[i].DosageText) ||
			!equalStringPointers(got[i].EndDate, want[i].EndDate) ||
			!equalStringPointers(got[i].Note, want[i].Note) {
			t.Fatalf("entry %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func equalStringPointers(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func stringPointer(value string) *string {
	return &value
}
