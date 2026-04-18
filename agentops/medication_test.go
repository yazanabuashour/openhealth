package agentops_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yazanabuashour/openhealth/agentops"
	"github.com/yazanabuashour/openhealth/client"
)

func TestRunMedicationTaskRecordListCorrectAndDelete(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	result, err := agentops.RunMedicationTask(ctx, config, agentops.MedicationTaskRequest{
		Action: agentops.MedicationTaskActionRecord,
		Medications: []agentops.MedicationInput{
			{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01"},
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
	assertMedicationEntries(t, result.Entries, []agentops.MedicationEntry{
		{Name: "Vitamin D", StartDate: "2026-02-01", EndDate: stringPointer("2026-03-01")},
		{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01"},
	})

	again, err := agentops.RunMedicationTask(ctx, config, agentops.MedicationTaskRequest{
		Action: agentops.MedicationTaskActionRecord,
		Medications: []agentops.MedicationInput{
			{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01"},
		},
	})
	if err != nil {
		t.Fatalf("repeat medication task: %v", err)
	}
	if got := medicationWriteStatuses(again.Writes); got != "already_exists" {
		t.Fatalf("repeat write statuses = %q, want already_exists", got)
	}

	conflict, err := agentops.RunMedicationTask(ctx, config, agentops.MedicationTaskRequest{
		Action: agentops.MedicationTaskActionRecord,
		Medications: []agentops.MedicationInput{
			{Name: "Levothyroxine", DosageText: stringPointer("50 mcg"), StartDate: "2026-01-01"},
		},
	})
	if err != nil {
		t.Fatalf("conflict medication task: %v", err)
	}
	if !conflict.Rejected || conflict.RejectionReason != "medication already exists for Levothyroxine starting 2026-01-01; use correct_medication" {
		t.Fatalf("conflict result = %#v", conflict)
	}

	corrected, err := agentops.RunMedicationTask(ctx, config, agentops.MedicationTaskRequest{
		Action: agentops.MedicationTaskActionCorrect,
		Target: &agentops.MedicationTarget{Name: "Levothyroxine", StartDate: "2026-01-01"},
		Medication: &agentops.MedicationInput{
			Name:       "Levothyroxine",
			DosageText: stringPointer("50 mcg"),
			StartDate:  "2026-01-01",
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
	assertMedicationEntries(t, corrected.Entries, []agentops.MedicationEntry{
		{Name: "Vitamin D", StartDate: "2026-02-01", EndDate: stringPointer("2026-03-01")},
		{Name: "Levothyroxine", DosageText: stringPointer("50 mcg"), StartDate: "2026-01-01"},
	})

	deleted, err := agentops.RunMedicationTask(ctx, config, agentops.MedicationTaskRequest{
		Action: agentops.MedicationTaskActionDelete,
		Target: &agentops.MedicationTarget{Name: "Vitamin D", StartDate: "2026-02-01"},
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
	assertMedicationEntries(t, deleted.Entries, []agentops.MedicationEntry{
		{Name: "Levothyroxine", DosageText: stringPointer("50 mcg"), StartDate: "2026-01-01"},
	})

	active, err := agentops.RunMedicationTask(ctx, config, agentops.MedicationTaskRequest{
		Action: agentops.MedicationTaskActionList,
		Status: agentops.MedicationStatusActive,
	})
	if err != nil {
		t.Fatalf("list active medications: %v", err)
	}
	assertMedicationEntries(t, active.Entries, []agentops.MedicationEntry{
		{Name: "Levothyroxine", DosageText: stringPointer("50 mcg"), StartDate: "2026-01-01"},
	})
}

func TestRunMedicationTaskRejectsMissingOrAmbiguousTarget(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	missing, err := agentops.RunMedicationTask(ctx, config, agentops.MedicationTaskRequest{
		Action: agentops.MedicationTaskActionCorrect,
		Target: &agentops.MedicationTarget{Name: "Levothyroxine", StartDate: "2026-01-01"},
		Medication: &agentops.MedicationInput{
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

	ambiguous, err := agentops.RunMedicationTask(ctx, config, agentops.MedicationTaskRequest{
		Action: agentops.MedicationTaskActionDelete,
		Target: &agentops.MedicationTarget{Name: "Levothyroxine", StartDate: "2026-01-01"},
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
		request agentops.MedicationTaskRequest
		reason  string
	}{
		{
			name: "short start date",
			request: agentops.MedicationTaskRequest{
				Action:      agentops.MedicationTaskActionRecord,
				Medications: []agentops.MedicationInput{{Name: "Levothyroxine", StartDate: "03/29"}},
			},
			reason: "start_date must be YYYY-MM-DD",
		},
		{
			name: "empty dosage",
			request: agentops.MedicationTaskRequest{
				Action:      agentops.MedicationTaskActionRecord,
				Medications: []agentops.MedicationInput{{Name: "Levothyroxine", DosageText: stringPointer(" "), StartDate: "2026-01-01"}},
			},
			reason: "dosage_text must not be empty",
		},
		{
			name: "end before start",
			request: agentops.MedicationTaskRequest{
				Action:      agentops.MedicationTaskActionRecord,
				Medications: []agentops.MedicationInput{{Name: "Levothyroxine", StartDate: "2026-01-02", EndDate: stringPointer("2026-01-01")}},
			},
			reason: "end_date must be on or after start_date",
		},
		{
			name: "unsupported status",
			request: agentops.MedicationTaskRequest{
				Action: agentops.MedicationTaskActionList,
				Status: "completed",
			},
			reason: "status must be active or all",
		},
		{
			name: "missing target",
			request: agentops.MedicationTaskRequest{
				Action:     agentops.MedicationTaskActionDelete,
				Target:     &agentops.MedicationTarget{Name: "Levothyroxine"},
				Medication: &agentops.MedicationInput{Name: "Levothyroxine", StartDate: "2026-01-01"},
			},
			reason: "target id or name and start_date are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dbPath := filepath.Join(t.TempDir(), "openhealth.db")
			result, err := agentops.RunMedicationTask(context.Background(), client.LocalConfig{DatabasePath: dbPath}, tt.request)
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

func medicationWriteStatuses(writes []agentops.MedicationWrite) string {
	out := ""
	for i, write := range writes {
		if i > 0 {
			out += ","
		}
		out += write.Status
	}
	return out
}

func assertMedicationEntries(t *testing.T, got []agentops.MedicationEntry, want []agentops.MedicationEntry) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("entry count = %d (%#v), want %d (%#v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i].Name != want[i].Name ||
			got[i].StartDate != want[i].StartDate ||
			!equalStringPointers(got[i].DosageText, want[i].DosageText) ||
			!equalStringPointers(got[i].EndDate, want[i].EndDate) {
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
