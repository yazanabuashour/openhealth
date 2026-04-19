package runner_test

import (
	"context"
	"path/filepath"
	"slices"
	"testing"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
	runner "github.com/yazanabuashour/openhealth/internal/runner"
)

func TestRunImagingTaskRecordListCorrectAndDelete(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	result, err := runner.RunImagingTask(ctx, config, runner.ImagingTaskRequest{
		Action: runner.ImagingTaskActionRecord,
		Records: []runner.ImagingInput{{
			Date:       "2026-03-29",
			Modality:   "X-ray",
			BodySite:   stringPointer("Chest"),
			Title:      stringPointer("Chest X-ray"),
			Summary:    "No acute cardiopulmonary abnormality.",
			Impression: stringPointer("Normal chest radiograph."),
			Note:       stringPointer("Imported from scan summary"),
			Notes:      []string{" XR TOE RIGHT narrative\nsecond line ", "Compared with prior study."},
		}},
	})
	if err != nil {
		t.Fatalf("record imaging: %v", err)
	}
	if result.Rejected {
		t.Fatalf("record rejected: %#v", result)
	}
	if got := imagingWriteStatuses(result.Writes); got != "created" {
		t.Fatalf("write status = %q, want created", got)
	}
	assertImagingEntries(t, result.Entries, []runner.ImagingTaskEntry{{
		Date:       "2026-03-29",
		Modality:   "X-ray",
		BodySite:   stringPointer("Chest"),
		Title:      stringPointer("Chest X-ray"),
		Summary:    "No acute cardiopulmonary abnormality.",
		Impression: stringPointer("Normal chest radiograph."),
		Note:       stringPointer("Imported from scan summary"),
		Notes:      []string{"XR TOE RIGHT narrative\nsecond line", "Compared with prior study."},
	}})

	again, err := runner.RunImagingTask(ctx, config, runner.ImagingTaskRequest{
		Action: runner.ImagingTaskActionRecord,
		Records: []runner.ImagingInput{{
			Date:       "2026-03-29",
			Modality:   "x-ray",
			BodySite:   stringPointer("chest"),
			Title:      stringPointer("Chest X-ray"),
			Summary:    "No acute cardiopulmonary abnormality.",
			Impression: stringPointer("Normal chest radiograph."),
			Note:       stringPointer("Imported from scan summary"),
			Notes:      []string{"XR TOE RIGHT narrative\nsecond line", "Compared with prior study."},
		}},
	})
	if err != nil {
		t.Fatalf("repeat imaging: %v", err)
	}
	if got := imagingWriteStatuses(again.Writes); got != "already_exists" {
		t.Fatalf("repeat status = %q, want already_exists", got)
	}

	second, err := runner.RunImagingTask(ctx, config, runner.ImagingTaskRequest{
		Action: runner.ImagingTaskActionRecord,
		Records: []runner.ImagingInput{{
			Date:     "2026-03-29",
			Modality: "MRI",
			BodySite: stringPointer("Knee"),
			Summary:  "Mild patellar tendinosis.",
		}},
	})
	if err != nil {
		t.Fatalf("second imaging: %v", err)
	}
	if got := imagingWriteStatuses(second.Writes); got != "created" {
		t.Fatalf("second status = %q, want created", got)
	}

	filtered, err := runner.RunImagingTask(ctx, config, runner.ImagingTaskRequest{
		Action:   runner.ImagingTaskActionList,
		ListMode: runner.ImagingListModeHistory,
		Modality: "mri",
		BodySite: "knee",
	})
	if err != nil {
		t.Fatalf("filtered imaging: %v", err)
	}
	assertImagingEntries(t, filtered.Entries, []runner.ImagingTaskEntry{{
		Date:     "2026-03-29",
		Modality: "MRI",
		BodySite: stringPointer("Knee"),
		Summary:  "Mild patellar tendinosis.",
	}})

	ambiguous, err := runner.RunImagingTask(ctx, config, runner.ImagingTaskRequest{
		Action: runner.ImagingTaskActionDelete,
		Target: &runner.ImagingTarget{Date: "2026-03-29"},
	})
	if err != nil {
		t.Fatalf("ambiguous delete: %v", err)
	}
	if !ambiguous.Rejected || ambiguous.RejectionReason != "multiple matching imaging records; target is ambiguous" {
		t.Fatalf("ambiguous result = %#v", ambiguous)
	}

	idTarget := result.Writes[0].ID
	corrected, err := runner.RunImagingTask(ctx, config, runner.ImagingTaskRequest{
		Action: runner.ImagingTaskActionCorrect,
		Target: &runner.ImagingTarget{ID: idTarget},
		Record: &runner.ImagingInput{
			Date:       "2026-03-30",
			Modality:   "MRI",
			BodySite:   stringPointer("Knee"),
			Summary:    "Mild patellar tendinosis, no tear.",
			Impression: stringPointer("No meniscal tear."),
			Notes:      []string{"MRI narrative"},
		},
	})
	if err != nil {
		t.Fatalf("correct imaging: %v", err)
	}
	if got := imagingWriteStatuses(corrected.Writes); got != "updated" {
		t.Fatalf("correct status = %q, want updated", got)
	}
	var correctedEntry *runner.ImagingTaskEntry
	for index := range corrected.Entries {
		if corrected.Entries[index].ID == idTarget {
			correctedEntry = &corrected.Entries[index]
			break
		}
	}
	if correctedEntry == nil || correctedEntry.Title == nil || *correctedEntry.Title != "Chest X-ray" || correctedEntry.Note == nil || *correctedEntry.Note != "Imported from scan summary" {
		t.Fatalf("corrected entry = %#v, want preserved title and note", correctedEntry)
	}
	if !slices.Equal(correctedEntry.Notes, []string{"MRI narrative"}) {
		t.Fatalf("corrected notes = %#v, want replacement imaging notes", correctedEntry.Notes)
	}

	correctedWithoutNotes, err := runner.RunImagingTask(ctx, config, runner.ImagingTaskRequest{
		Action: runner.ImagingTaskActionCorrect,
		Target: &runner.ImagingTarget{ID: idTarget},
		Record: &runner.ImagingInput{
			Date:     "2026-03-30",
			Modality: "MRI",
			BodySite: stringPointer("Knee"),
			Summary:  "Mild patellar tendinosis without tear.",
		},
	})
	if err != nil {
		t.Fatalf("correct imaging without notes: %v", err)
	}
	var preservedNotesEntry *runner.ImagingTaskEntry
	for index := range correctedWithoutNotes.Entries {
		if correctedWithoutNotes.Entries[index].ID == idTarget {
			preservedNotesEntry = &correctedWithoutNotes.Entries[index]
			break
		}
	}
	if preservedNotesEntry == nil || !slices.Equal(preservedNotesEntry.Notes, []string{"MRI narrative"}) {
		t.Fatalf("corrected notes after omitted notes = %#v, want preserved MRI narrative", preservedNotesEntry)
	}

	deleted, err := runner.RunImagingTask(ctx, config, runner.ImagingTaskRequest{
		Action: runner.ImagingTaskActionDelete,
		Target: &runner.ImagingTarget{ID: idTarget},
	})
	if err != nil {
		t.Fatalf("delete imaging: %v", err)
	}
	if got := imagingWriteStatuses(deleted.Writes); got != "deleted" {
		t.Fatalf("delete status = %q, want deleted", got)
	}
}

func TestRunImagingTaskRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  runner.ImagingInput
		reason string
	}{
		{name: "missing modality", input: runner.ImagingInput{Date: "2026-03-29", Summary: "Normal."}, reason: "modality is required"},
		{name: "missing summary", input: runner.ImagingInput{Date: "2026-03-29", Modality: "X-ray"}, reason: "summary is required"},
		{name: "empty note", input: runner.ImagingInput{Date: "2026-03-29", Modality: "X-ray", Summary: "Normal.", Note: stringPointer(" ")}, reason: "note must not be empty"},
		{name: "empty notes item", input: runner.ImagingInput{Date: "2026-03-29", Modality: "X-ray", Summary: "Normal.", Notes: []string{"finding", " "}}, reason: "notes must not contain empty values"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := runner.RunImagingTask(context.Background(), client.LocalConfig{DatabasePath: filepath.Join(t.TempDir(), "openhealth.db")}, runner.ImagingTaskRequest{
				Action:  runner.ImagingTaskActionRecord,
				Records: []runner.ImagingInput{tt.input},
			})
			if err != nil {
				t.Fatalf("run imaging: %v", err)
			}
			if !result.Rejected || result.RejectionReason != tt.reason {
				t.Fatalf("result = %#v, want rejection %q", result, tt.reason)
			}
		})
	}
}

func imagingWriteStatuses(writes []runner.ImagingWrite) string {
	out := ""
	for i, write := range writes {
		if i > 0 {
			out += ","
		}
		out += write.Status
	}
	return out
}

func assertImagingEntries(t *testing.T, got []runner.ImagingTaskEntry, want []runner.ImagingTaskEntry) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("entry count = %d (%#v), want %d (%#v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i].Date != want[i].Date ||
			got[i].Modality != want[i].Modality ||
			got[i].Summary != want[i].Summary ||
			!equalStringPointers(got[i].BodySite, want[i].BodySite) ||
			!equalStringPointers(got[i].Title, want[i].Title) ||
			!equalStringPointers(got[i].Impression, want[i].Impression) ||
			!equalStringPointers(got[i].Note, want[i].Note) ||
			!slices.Equal(got[i].Notes, want[i].Notes) {
			t.Fatalf("entry %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}
