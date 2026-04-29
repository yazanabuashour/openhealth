package runner_test

import (
	"context"
	"path/filepath"
	"testing"

	client "github.com/yazanabuashour/openhealth/client"
	runner "github.com/yazanabuashour/openhealth/internal/runner"
)

func TestRunBodyCompositionTaskRecordListCorrectAndDelete(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	result, err := runner.RunBodyCompositionTask(ctx, config, runner.BodyCompositionTaskRequest{
		Action: runner.BodyCompositionTaskActionRecord,
		Records: []runner.BodyCompositionInput{
			{
				Date:           "2026-03-29",
				BodyFatPercent: floatPointer(18.7),
				WeightValue:    floatPointer(154.2),
				WeightUnit:     stringPointer("lbs"),
				Method:         stringPointer("smart scale"),
				Note:           stringPointer("Imported from scale report"),
			},
		},
	})
	if err != nil {
		t.Fatalf("record body composition: %v", err)
	}
	if result.Rejected {
		t.Fatalf("record rejected: %#v", result)
	}
	if got := bodyCompositionWriteStatuses(result.Writes); got != "created" {
		t.Fatalf("write statuses = %q, want created", got)
	}
	assertBodyCompositionEntries(t, result.Entries, []runner.BodyCompositionTaskEntry{{
		Date:           "2026-03-29",
		BodyFatPercent: floatPointer(18.7),
		WeightValue:    floatPointer(154.2),
		WeightUnit:     stringPointer("lb"),
		Method:         stringPointer("smart scale"),
		Note:           stringPointer("Imported from scale report"),
	}})

	again, err := runner.RunBodyCompositionTask(ctx, config, runner.BodyCompositionTaskRequest{
		Action: runner.BodyCompositionTaskActionRecord,
		Records: []runner.BodyCompositionInput{{
			Date:           "2026-03-29",
			BodyFatPercent: floatPointer(18.7),
			WeightValue:    floatPointer(154.2),
			WeightUnit:     stringPointer("lb"),
			Method:         stringPointer("smart scale"),
			Note:           stringPointer("Imported from scale report"),
		}},
	})
	if err != nil {
		t.Fatalf("repeat body composition: %v", err)
	}
	if got := bodyCompositionWriteStatuses(again.Writes); got != "already_exists" {
		t.Fatalf("repeat status = %q, want already_exists", got)
	}

	second, err := runner.RunBodyCompositionTask(ctx, config, runner.BodyCompositionTaskRequest{
		Action: runner.BodyCompositionTaskActionRecord,
		Records: []runner.BodyCompositionInput{{
			Date:           "2026-03-29",
			BodyFatPercent: floatPointer(18.5),
		}},
	})
	if err != nil {
		t.Fatalf("second same-day body composition: %v", err)
	}
	if got := bodyCompositionWriteStatuses(second.Writes); got != "created" {
		t.Fatalf("second status = %q, want created", got)
	}

	ambiguous, err := runner.RunBodyCompositionTask(ctx, config, runner.BodyCompositionTaskRequest{
		Action: runner.BodyCompositionTaskActionDelete,
		Target: &runner.BodyCompositionTarget{Date: "2026-03-29"},
	})
	if err != nil {
		t.Fatalf("ambiguous delete: %v", err)
	}
	if !ambiguous.Rejected || ambiguous.RejectionReason != "multiple matching body composition entries; target is ambiguous" {
		t.Fatalf("ambiguous result = %#v", ambiguous)
	}

	idTarget := result.Writes[0].ID
	corrected, err := runner.RunBodyCompositionTask(ctx, config, runner.BodyCompositionTaskRequest{
		Action: runner.BodyCompositionTaskActionCorrect,
		Target: &runner.BodyCompositionTarget{ID: idTarget},
		Record: &runner.BodyCompositionInput{
			Date:           "2026-03-30",
			BodyFatPercent: floatPointer(18.4),
			Method:         stringPointer("smart scale"),
		},
	})
	if err != nil {
		t.Fatalf("correct body composition: %v", err)
	}
	if got := bodyCompositionWriteStatuses(corrected.Writes); got != "updated" {
		t.Fatalf("correct status = %q, want updated", got)
	}
	var correctedEntry *runner.BodyCompositionTaskEntry
	for index := range corrected.Entries {
		if corrected.Entries[index].ID == idTarget {
			correctedEntry = &corrected.Entries[index]
			break
		}
	}
	if correctedEntry == nil || correctedEntry.Note == nil || *correctedEntry.Note != "Imported from scale report" {
		t.Fatalf("corrected entry = %#v, want preserved note", correctedEntry)
	}

	deleted, err := runner.RunBodyCompositionTask(ctx, config, runner.BodyCompositionTaskRequest{
		Action: runner.BodyCompositionTaskActionDelete,
		Target: &runner.BodyCompositionTarget{ID: idTarget},
	})
	if err != nil {
		t.Fatalf("delete body composition: %v", err)
	}
	if got := bodyCompositionWriteStatuses(deleted.Writes); got != "deleted" {
		t.Fatalf("delete status = %q, want deleted", got)
	}
}

func TestRunBodyCompositionTaskRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  runner.BodyCompositionInput
		reason string
	}{
		{name: "missing measurement", input: runner.BodyCompositionInput{Date: "2026-03-29"}, reason: "at least one body composition measurement is required"},
		{name: "invalid body fat", input: runner.BodyCompositionInput{Date: "2026-03-29", BodyFatPercent: floatPointer(101)}, reason: "body_fat_percent must be greater than 0 and less than or equal to 100"},
		{name: "missing weight unit", input: runner.BodyCompositionInput{Date: "2026-03-29", WeightValue: floatPointer(154.2)}, reason: "weight_value and weight_unit must be provided together"},
		{name: "unsupported weight unit", input: runner.BodyCompositionInput{Date: "2026-03-29", WeightValue: floatPointer(154.2), WeightUnit: stringPointer("kg")}, reason: "weight_unit must be lb"},
		{name: "empty note", input: runner.BodyCompositionInput{Date: "2026-03-29", BodyFatPercent: floatPointer(18.7), Note: stringPointer(" ")}, reason: "note must not be empty"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := runner.RunBodyCompositionTask(context.Background(), client.LocalConfig{DatabasePath: filepath.Join(t.TempDir(), "openhealth.db")}, runner.BodyCompositionTaskRequest{
				Action:  runner.BodyCompositionTaskActionRecord,
				Records: []runner.BodyCompositionInput{tt.input},
			})
			if err != nil {
				t.Fatalf("run task: %v", err)
			}
			if !result.Rejected || result.RejectionReason != tt.reason {
				t.Fatalf("result = %#v, want rejection %q", result, tt.reason)
			}
		})
	}
}

func bodyCompositionWriteStatuses(writes []runner.BodyCompositionWrite) string {
	out := ""
	for i, write := range writes {
		if i > 0 {
			out += ","
		}
		out += write.Status
	}
	return out
}

func assertBodyCompositionEntries(t *testing.T, got []runner.BodyCompositionTaskEntry, want []runner.BodyCompositionTaskEntry) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("entry count = %d (%#v), want %d (%#v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i].Date != want[i].Date ||
			!equalFloatPointers(got[i].BodyFatPercent, want[i].BodyFatPercent) ||
			!equalFloatPointers(got[i].WeightValue, want[i].WeightValue) ||
			!equalStringPointers(got[i].WeightUnit, want[i].WeightUnit) ||
			!equalStringPointers(got[i].Method, want[i].Method) ||
			!equalStringPointers(got[i].Note, want[i].Note) {
			t.Fatalf("entry %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func equalFloatPointers(left *float64, right *float64) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}
