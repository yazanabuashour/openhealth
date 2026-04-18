package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
	runner "github.com/yazanabuashour/openhealth/internal/runner"
)

func TestRunLabTaskRecordListCorrectAndDelete(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	result, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action: runner.LabTaskActionRecord,
		Collections: []runner.LabCollectionInput{
			metabolicCollection("2026-03-29", 89),
			thyroidCollection("2026-03-30", "3.1"),
		},
	})
	if err != nil {
		t.Fatalf("run record task: %v", err)
	}
	if result.Rejected {
		t.Fatalf("result rejected: %#v", result)
	}
	if got := labWriteStatuses(result.Writes); got != "created,created" {
		t.Fatalf("write statuses = %q, want created,created", got)
	}
	assertLabCollectionDates(t, result.Entries, []string{"2026-03-30", "2026-03-29"})

	again, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action:      runner.LabTaskActionRecord,
		Collections: []runner.LabCollectionInput{metabolicCollection("2026-03-29", 89)},
	})
	if err != nil {
		t.Fatalf("repeat lab task: %v", err)
	}
	if got := labWriteStatuses(again.Writes); got != "already_exists" {
		t.Fatalf("repeat write statuses = %q, want already_exists", got)
	}

	conflict, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action:      runner.LabTaskActionRecord,
		Collections: []runner.LabCollectionInput{metabolicCollection("2026-03-29", 90)},
	})
	if err != nil {
		t.Fatalf("conflict lab task: %v", err)
	}
	if !conflict.Rejected || conflict.RejectionReason != "lab collection already exists for 2026-03-29; use correct_labs" {
		t.Fatalf("conflict result = %#v", conflict)
	}

	latestGlucose, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action:      runner.LabTaskActionList,
		ListMode:    runner.LabListModeLatest,
		AnalyteSlug: "glucose",
	})
	if err != nil {
		t.Fatalf("latest glucose task: %v", err)
	}
	assertLabCollectionDates(t, latestGlucose.Entries, []string{"2026-03-29"})
	if got := latestGlucose.Entries[0].Panels[0].Results[0].ValueText; got != "89" {
		t.Fatalf("glucose value = %q, want 89", got)
	}

	bounded, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action:   runner.LabTaskActionList,
		ListMode: runner.LabListModeRange,
		FromDate: "2026-03-29",
		ToDate:   "2026-03-29",
	})
	if err != nil {
		t.Fatalf("bounded range task: %v", err)
	}
	assertLabCollectionDates(t, bounded.Entries, []string{"2026-03-29"})

	corrected, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action:     runner.LabTaskActionCorrect,
		Target:     &runner.LabTarget{Date: "2026-03-29"},
		Collection: &runner.LabCollectionInput{Date: "2026-03-29", Panels: []runner.LabPanelInput{{PanelName: "Thyroid", Results: []runner.LabResultInput{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "2.8", ValueNumeric: floatPointer(2.8), Units: stringPointer("uIU/mL")}}}}},
	})
	if err != nil {
		t.Fatalf("correct lab task: %v", err)
	}
	if corrected.Rejected {
		t.Fatalf("correction rejected: %#v", corrected)
	}
	if got := labWriteStatuses(corrected.Writes); got != "updated" {
		t.Fatalf("correction write statuses = %q, want updated", got)
	}
	correctedLatest, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action:      runner.LabTaskActionList,
		ListMode:    runner.LabListModeLatest,
		AnalyteSlug: "tsh",
	})
	if err != nil {
		t.Fatalf("list corrected tsh task: %v", err)
	}
	assertLabCollectionDates(t, correctedLatest.Entries, []string{"2026-03-30"})

	deleted, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action: runner.LabTaskActionDelete,
		Target: &runner.LabTarget{Date: "2026-03-29"},
	})
	if err != nil {
		t.Fatalf("delete lab task: %v", err)
	}
	if deleted.Rejected {
		t.Fatalf("deletion rejected: %#v", deleted)
	}
	if got := labWriteStatuses(deleted.Writes); got != "deleted" {
		t.Fatalf("delete write statuses = %q, want deleted", got)
	}
	assertLabCollectionDates(t, deleted.Entries, []string{"2026-03-30"})
}

func TestRunLabTaskRejectsMissingOrAmbiguousTarget(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	missing, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action:     runner.LabTaskActionCorrect,
		Target:     &runner.LabTarget{Date: "2026-03-29"},
		Collection: &runner.LabCollectionInput{Date: "2026-03-29", Panels: []runner.LabPanelInput{{PanelName: "Metabolic", Results: []runner.LabResultInput{{TestName: "Glucose", ValueText: "89"}}}}},
	})
	if err != nil {
		t.Fatalf("missing correction: %v", err)
	}
	if !missing.Rejected || missing.RejectionReason != "no matching lab collection" {
		t.Fatalf("missing result = %#v", missing)
	}

	api, err := client.OpenLocal(config)
	if err != nil {
		t.Fatalf("open local client: %v", err)
	}
	_, err = api.CreateLabCollection(ctx, client.LabCollectionInput{
		CollectedAt: mustDate(t, "2026-03-29"),
		Panels:      []client.LabPanelInput{{PanelName: "Metabolic", Results: []client.LabResultInput{{TestName: "Glucose", ValueText: "89"}}}},
	})
	if err != nil {
		t.Fatalf("create first lab: %v", err)
	}
	_, err = api.CreateLabCollection(ctx, client.LabCollectionInput{
		CollectedAt: mustDate(t, "2026-03-29"),
		Panels:      []client.LabPanelInput{{PanelName: "Thyroid", Results: []client.LabResultInput{{TestName: "TSH", ValueText: "3.1"}}}},
	})
	if err != nil {
		t.Fatalf("create second lab: %v", err)
	}
	if err := api.Close(); err != nil {
		t.Fatalf("close local client: %v", err)
	}

	ambiguous, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action: runner.LabTaskActionDelete,
		Target: &runner.LabTarget{Date: "2026-03-29"},
	})
	if err != nil {
		t.Fatalf("ambiguous delete: %v", err)
	}
	if !ambiguous.Rejected || ambiguous.RejectionReason != "multiple matching lab collections; target is ambiguous" {
		t.Fatalf("ambiguous result = %#v", ambiguous)
	}
}

func TestRunLabTaskRejectsInvalidInputBeforeOpeningDatabase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		request runner.LabTaskRequest
		reason  string
	}{
		{
			name: "short date",
			request: runner.LabTaskRequest{
				Action:      runner.LabTaskActionRecord,
				Collections: []runner.LabCollectionInput{{Date: "03/29", Panels: []runner.LabPanelInput{{PanelName: "Metabolic", Results: []runner.LabResultInput{{TestName: "Glucose", ValueText: "89"}}}}}},
			},
			reason: "date must be YYYY-MM-DD",
		},
		{
			name: "unsupported slug",
			request: runner.LabTaskRequest{
				Action:      runner.LabTaskActionRecord,
				Collections: []runner.LabCollectionInput{{Date: "2026-03-29", Panels: []runner.LabPanelInput{{PanelName: "Metabolic", Results: []runner.LabResultInput{{TestName: "Unknown", CanonicalSlug: stringPointer("unsupported"), ValueText: "1"}}}}}},
			},
			reason: "canonical_slug must be a supported analyte",
		},
		{
			name: "empty units",
			request: runner.LabTaskRequest{
				Action:      runner.LabTaskActionRecord,
				Collections: []runner.LabCollectionInput{{Date: "2026-03-29", Panels: []runner.LabPanelInput{{PanelName: "Metabolic", Results: []runner.LabResultInput{{TestName: "Glucose", ValueText: "89", Units: stringPointer(" ")}}}}}},
			},
			reason: "units must not be empty",
		},
		{
			name: "unsupported list analyte",
			request: runner.LabTaskRequest{
				Action:      runner.LabTaskActionList,
				ListMode:    runner.LabListModeLatest,
				AnalyteSlug: "unsupported",
			},
			reason: "analyte_slug must be a supported analyte",
		},
		{
			name: "missing target date",
			request: runner.LabTaskRequest{
				Action: runner.LabTaskActionDelete,
				Target: &runner.LabTarget{},
			},
			reason: "target id or date is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dbPath := filepath.Join(t.TempDir(), "openhealth.db")
			result, err := runner.RunLabTask(context.Background(), client.LocalConfig{DatabasePath: dbPath}, tt.request)
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

func metabolicCollection(date string, value float64) runner.LabCollectionInput {
	return runner.LabCollectionInput{
		Date: date,
		Panels: []runner.LabPanelInput{
			{
				PanelName: "Metabolic",
				Results: []runner.LabResultInput{
					{
						TestName:      "Glucose",
						CanonicalSlug: stringPointer("glucose"),
						ValueText:     trimFloat(value),
						ValueNumeric:  floatPointer(value),
						Units:         stringPointer("mg/dL"),
						RangeText:     stringPointer("70-99"),
					},
				},
			},
		},
	}
}

func thyroidCollection(date string, value string) runner.LabCollectionInput {
	return runner.LabCollectionInput{
		Date: date,
		Panels: []runner.LabPanelInput{
			{
				PanelName: "Thyroid",
				Results: []runner.LabResultInput{
					{
						TestName:      "TSH",
						CanonicalSlug: stringPointer("tsh"),
						ValueText:     value,
						Units:         stringPointer("uIU/mL"),
					},
				},
			},
		},
	}
}

func labWriteStatuses(writes []runner.LabCollectionWrite) string {
	out := ""
	for i, write := range writes {
		if i > 0 {
			out += ","
		}
		out += write.Status
	}
	return out
}

func assertLabCollectionDates(t *testing.T, got []runner.LabCollectionEntry, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("entry count = %d (%#v), want %d dates %#v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i].Date != want[i] {
			t.Fatalf("entry %d date = %q, want %q; entries %#v", i, got[i].Date, want[i], got)
		}
	}
}

func floatPointer(value float64) *float64 {
	return &value
}

func trimFloat(value float64) string {
	switch value {
	case 89:
		return "89"
	case 90:
		return "90"
	default:
		return "0"
	}
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		t.Fatalf("parse date %q: %v", value, err)
	}
	return parsed
}
