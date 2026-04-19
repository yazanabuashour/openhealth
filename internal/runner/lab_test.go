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

func TestRunLabTaskRecordsMultipleSameDayCollections(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	first, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action:      runner.LabTaskActionRecord,
		Collections: []runner.LabCollectionInput{metabolicCollection("2026-03-29", 89)},
	})
	if err != nil {
		t.Fatalf("record first same-day lab: %v", err)
	}
	if got := labWriteStatuses(first.Writes); got != "created" {
		t.Fatalf("first write status = %q, want created", got)
	}

	second, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action:      runner.LabTaskActionRecord,
		Collections: []runner.LabCollectionInput{thyroidCollection("2026-03-29", "3.1")},
	})
	if err != nil {
		t.Fatalf("record second same-day lab: %v", err)
	}
	if second.Rejected {
		t.Fatalf("second same-day lab rejected: %#v", second)
	}
	if got := labWriteStatuses(second.Writes); got != "created" {
		t.Fatalf("second write status = %q, want created", got)
	}
	assertLabCollectionDates(t, second.Entries, []string{"2026-03-29", "2026-03-29"})

	duplicate, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action:      runner.LabTaskActionRecord,
		Collections: []runner.LabCollectionInput{metabolicCollection("2026-03-29", 89)},
	})
	if err != nil {
		t.Fatalf("repeat same-day lab: %v", err)
	}
	if got := labWriteStatuses(duplicate.Writes); got != "already_exists" {
		t.Fatalf("repeat write status = %q, want already_exists", got)
	}
}

func TestRunLabTaskPatchUpdatesOneLabResult(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	recorded, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action:      runner.LabTaskActionRecord,
		Collections: []runner.LabCollectionInput{metabolicCollectionWithLipids("2026-03-29")},
	})
	if err != nil {
		t.Fatalf("seed lab: %v", err)
	}
	if len(recorded.Writes) != 1 {
		t.Fatalf("recorded writes = %#v, want one", recorded.Writes)
	}

	patched, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action: runner.LabTaskActionPatch,
		Target: &runner.LabTarget{ID: recorded.Writes[0].ID},
		ResultUpdates: []runner.LabResultUpdateInput{
			{
				PanelName: "metabolic",
				Match:     runner.LabResultMatchInput{CanonicalSlug: "glucose"},
				Result: runner.LabResultInput{
					TestName:      "Glucose",
					CanonicalSlug: stringPointer("glucose"),
					ValueText:     "92",
					ValueNumeric:  floatPointer(92),
					Units:         stringPointer("mg/dL"),
					RangeText:     stringPointer("70-99"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("patch lab: %v", err)
	}
	if patched.Rejected {
		t.Fatalf("patch rejected: %#v", patched)
	}
	if got := labWriteStatuses(patched.Writes); got != "updated" {
		t.Fatalf("patch write status = %q, want updated", got)
	}
	assertLabCollectionDates(t, patched.Entries, []string{"2026-03-29"})
	results := patched.Entries[0].Panels[0].Results
	if len(results) != 2 {
		t.Fatalf("patched results = %#v, want two results", results)
	}
	assertLabResult(t, results[0], "Glucose", "glucose", "92")
	assertLabResult(t, results[1], "HDL", "hdl", "51")
}

func TestRunLabTaskPatchRejectsAmbiguousDateTarget(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	_, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action: runner.LabTaskActionRecord,
		Collections: []runner.LabCollectionInput{
			metabolicCollection("2026-03-29", 89),
			thyroidCollection("2026-03-29", "3.1"),
		},
	})
	if err != nil {
		t.Fatalf("seed same-day labs: %v", err)
	}

	patched, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action: runner.LabTaskActionPatch,
		Target: &runner.LabTarget{Date: "2026-03-29"},
		ResultUpdates: []runner.LabResultUpdateInput{
			{
				PanelName: "Metabolic",
				Match:     runner.LabResultMatchInput{CanonicalSlug: "glucose"},
				Result:    runner.LabResultInput{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "92"},
			},
		},
	})
	if err != nil {
		t.Fatalf("patch ambiguous same-day target: %v", err)
	}
	if !patched.Rejected || patched.RejectionReason != "multiple matching lab collections; target is ambiguous" {
		t.Fatalf("patch result = %#v", patched)
	}
}

func TestRunLabTaskPatchRejectsUnsafeMatches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		collection runner.LabCollectionInput
		update     runner.LabResultUpdateInput
		reason     string
	}{
		{
			name:       "missing panel",
			collection: metabolicCollectionWithLipids("2026-03-29"),
			update: runner.LabResultUpdateInput{
				PanelName: "Thyroid",
				Match:     runner.LabResultMatchInput{CanonicalSlug: "glucose"},
				Result:    runner.LabResultInput{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "92"},
			},
			reason: "no matching lab panel",
		},
		{
			name:       "duplicate panels",
			collection: duplicatePanelCollection("2026-03-29"),
			update: runner.LabResultUpdateInput{
				PanelName: "Metabolic",
				Match:     runner.LabResultMatchInput{CanonicalSlug: "glucose"},
				Result:    runner.LabResultInput{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "92"},
			},
			reason: "multiple matching lab panels; patch is ambiguous",
		},
		{
			name:       "duplicate results",
			collection: duplicateResultCollection("2026-03-29"),
			update: runner.LabResultUpdateInput{
				PanelName: "Metabolic",
				Match:     runner.LabResultMatchInput{CanonicalSlug: "glucose"},
				Result:    runner.LabResultInput{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "92"},
			},
			reason: "multiple matching lab results; patch is ambiguous",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dbPath := filepath.Join(t.TempDir(), "openhealth.db")
			ctx := context.Background()
			config := client.LocalConfig{DatabasePath: dbPath}

			recorded, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
				Action:      runner.LabTaskActionRecord,
				Collections: []runner.LabCollectionInput{tt.collection},
			})
			if err != nil {
				t.Fatalf("seed lab: %v", err)
			}
			patched, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
				Action:        runner.LabTaskActionPatch,
				Target:        &runner.LabTarget{ID: recorded.Writes[0].ID},
				ResultUpdates: []runner.LabResultUpdateInput{tt.update},
			})
			if err != nil {
				t.Fatalf("patch lab: %v", err)
			}
			if !patched.Rejected || patched.RejectionReason != tt.reason {
				t.Fatalf("patch result = %#v, want rejection %q", patched, tt.reason)
			}

			listed, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
				Action: runner.LabTaskActionList,
			})
			if err != nil {
				t.Fatalf("list after rejected patch: %v", err)
			}
			if firstValueText(listed.Entries) == "92" {
				t.Fatalf("rejected patch mutated collection: %#v", listed.Entries)
			}
		})
	}
}

func TestRunLabTaskAcceptsArbitraryLabSlugs(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "openhealth.db")
	ctx := context.Background()
	config := client.LocalConfig{DatabasePath: dbPath}

	result, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action: runner.LabTaskActionRecord,
		Collections: []runner.LabCollectionInput{
			arbitraryLabCollection("2026-04-01"),
			thyroidCollection("2026-04-02", "2.9"),
		},
	})
	if err != nil {
		t.Fatalf("record arbitrary labs: %v", err)
	}
	if result.Rejected {
		t.Fatalf("result rejected: %#v", result)
	}

	vitaminD, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action:      runner.LabTaskActionList,
		ListMode:    runner.LabListModeLatest,
		AnalyteSlug: "Vitamin D",
	})
	if err != nil {
		t.Fatalf("list vitamin d: %v", err)
	}
	assertLabCollectionDates(t, vitaminD.Entries, []string{"2026-04-01"})
	assertSingleLabResult(t, vitaminD.Entries[0], "Vitamin D", "vitamin-d", "32")

	hemoglobinA1C, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action:      runner.LabTaskActionList,
		ListMode:    runner.LabListModeLatest,
		AnalyteSlug: "hemoglobin_a1c",
	})
	if err != nil {
		t.Fatalf("list hemoglobin a1c: %v", err)
	}
	assertLabCollectionDates(t, hemoglobinA1C.Entries, []string{"2026-04-01"})
	assertSingleLabResult(t, hemoglobinA1C.Entries[0], "Hemoglobin A1c", "hemoglobin-a1c", "5.4")

	uacr, err := runner.RunLabTask(ctx, config, runner.LabTaskRequest{
		Action:      runner.LabTaskActionList,
		ListMode:    runner.LabListModeLatest,
		AnalyteSlug: "urine-albumin-creatinine-ratio",
	})
	if err != nil {
		t.Fatalf("list uacr: %v", err)
	}
	assertLabCollectionDates(t, uacr.Entries, []string{"2026-04-01"})
	assertSingleLabResult(t, uacr.Entries[0], "Urine Albumin Creatinine Ratio", "urine-albumin-creatinine-ratio", "9")
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
			name: "empty canonical slug",
			request: runner.LabTaskRequest{
				Action:      runner.LabTaskActionRecord,
				Collections: []runner.LabCollectionInput{{Date: "2026-03-29", Panels: []runner.LabPanelInput{{PanelName: "Metabolic", Results: []runner.LabResultInput{{TestName: "Unknown", CanonicalSlug: stringPointer(" "), ValueText: "1"}}}}}},
			},
			reason: "canonical_slug must be a valid analyte slug",
		},
		{
			name: "invalid canonical slug shape",
			request: runner.LabTaskRequest{
				Action:      runner.LabTaskActionRecord,
				Collections: []runner.LabCollectionInput{{Date: "2026-03-29", Panels: []runner.LabPanelInput{{PanelName: "Metabolic", Results: []runner.LabResultInput{{TestName: "Unknown", CanonicalSlug: stringPointer("vitamin/d"), ValueText: "1"}}}}}},
			},
			reason: "canonical_slug must be a valid analyte slug",
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
			name: "invalid list analyte shape",
			request: runner.LabTaskRequest{
				Action:      runner.LabTaskActionList,
				ListMode:    runner.LabListModeLatest,
				AnalyteSlug: "!!!",
			},
			reason: "analyte_slug must be a valid analyte slug",
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

func metabolicCollectionWithLipids(date string) runner.LabCollectionInput {
	return runner.LabCollectionInput{
		Date: date,
		Panels: []runner.LabPanelInput{
			{
				PanelName: "Metabolic",
				Results: []runner.LabResultInput{
					{
						TestName:      "Glucose",
						CanonicalSlug: stringPointer("glucose"),
						ValueText:     "89",
						ValueNumeric:  floatPointer(89),
						Units:         stringPointer("mg/dL"),
						RangeText:     stringPointer("70-99"),
					},
					{
						TestName:      "HDL",
						CanonicalSlug: stringPointer("hdl"),
						ValueText:     "51",
						ValueNumeric:  floatPointer(51),
						Units:         stringPointer("mg/dL"),
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

func duplicatePanelCollection(date string) runner.LabCollectionInput {
	return runner.LabCollectionInput{
		Date: date,
		Panels: []runner.LabPanelInput{
			{
				PanelName: "Metabolic",
				Results:   []runner.LabResultInput{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89"}},
			},
			{
				PanelName: "metabolic",
				Results:   []runner.LabResultInput{{TestName: "HDL", CanonicalSlug: stringPointer("hdl"), ValueText: "51"}},
			},
		},
	}
}

func duplicateResultCollection(date string) runner.LabCollectionInput {
	return runner.LabCollectionInput{
		Date: date,
		Panels: []runner.LabPanelInput{
			{
				PanelName: "Metabolic",
				Results: []runner.LabResultInput{
					{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89"},
					{TestName: "Glucose Repeat", CanonicalSlug: stringPointer("glucose"), ValueText: "90"},
				},
			},
		},
	}
}

func arbitraryLabCollection(date string) runner.LabCollectionInput {
	return runner.LabCollectionInput{
		Date: date,
		Panels: []runner.LabPanelInput{
			{
				PanelName: "Micronutrients",
				Results: []runner.LabResultInput{
					{
						TestName:      "Vitamin D",
						CanonicalSlug: stringPointer("Vitamin D"),
						ValueText:     "32",
						ValueNumeric:  floatPointer(32),
						Units:         stringPointer("ng/mL"),
					},
					{
						TestName:      "Hemoglobin A1c",
						CanonicalSlug: stringPointer("hemoglobin_a1c"),
						ValueText:     "5.4",
						ValueNumeric:  floatPointer(5.4),
						Units:         stringPointer("%"),
					},
					{
						TestName:      "Urine Albumin Creatinine Ratio",
						CanonicalSlug: stringPointer("urine  albumin_creatinine   ratio"),
						ValueText:     "9",
						ValueNumeric:  floatPointer(9),
						Units:         stringPointer("mg/g"),
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

func assertSingleLabResult(t *testing.T, entry runner.LabCollectionEntry, testName string, slug string, valueText string) {
	t.Helper()
	if len(entry.Panels) != 1 || len(entry.Panels[0].Results) != 1 {
		t.Fatalf("entry = %#v, want exactly one nested result", entry)
	}
	assertLabResult(t, entry.Panels[0].Results[0], testName, slug, valueText)
}

func assertLabResult(t *testing.T, result runner.LabResultEntry, testName string, slug string, valueText string) {
	t.Helper()
	if result.TestName != testName || result.CanonicalSlug == nil || *result.CanonicalSlug != slug || result.ValueText != valueText {
		t.Fatalf("result = %#v, want %s %s %s", result, testName, slug, valueText)
	}
}

func firstValueText(entries []runner.LabCollectionEntry) string {
	for _, entry := range entries {
		for _, panel := range entry.Panels {
			for _, result := range panel.Results {
				return result.ValueText
			}
		}
	}
	return ""
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
