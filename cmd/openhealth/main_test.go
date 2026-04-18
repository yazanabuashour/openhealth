package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	runner "github.com/yazanabuashour/openhealth/internal/runner"
)

func TestRunWeightJSONRoundTrip(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "nested", "openhealth.db")

	upsert := `{"action":"upsert_weights","weights":[{"date":"2026-03-29","value":152.2,"unit":"lb"},{"date":"2026-03-30","value":151.6,"unit":"lb"}]}`
	var upsertStdout bytes.Buffer
	if err := run([]string{"weight", "--db", databasePath}, strings.NewReader(upsert), &upsertStdout, ioDiscard{}); err != nil {
		t.Fatalf("run weight upsert: %v", err)
	}
	var upsertResult runner.WeightTaskResult
	decodeJSON(t, upsertStdout.Bytes(), &upsertResult)
	if len(upsertResult.Writes) != 2 {
		t.Fatalf("writes = %d, want 2", len(upsertResult.Writes))
	}
	assertWeightEntries(t, upsertResult.Entries, []runner.WeightTaskEntry{
		{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
	})

	list := `{"action":"list_weights","list_mode":"history","limit":2}`
	var listStdout bytes.Buffer
	if err := run([]string{"weight", "--db", databasePath}, strings.NewReader(list), &listStdout, ioDiscard{}); err != nil {
		t.Fatalf("run weight list: %v", err)
	}
	var listResult runner.WeightTaskResult
	decodeJSON(t, listStdout.Bytes(), &listResult)
	assertWeightEntries(t, listResult.Entries, []runner.WeightTaskEntry{
		{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
	})
}

func TestRunBloodPressureJSONRoundTrip(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "nested", "openhealth.db")

	record := `{"action":"record_blood_pressure","readings":[{"date":"2026-03-29","systolic":122,"diastolic":78,"pulse":64},{"date":"2026-03-30","systolic":118,"diastolic":76}]}`
	var recordStdout bytes.Buffer
	if err := run([]string{"blood-pressure", "--db", databasePath}, strings.NewReader(record), &recordStdout, ioDiscard{}); err != nil {
		t.Fatalf("run blood-pressure record: %v", err)
	}
	var recordResult runner.BloodPressureTaskResult
	decodeJSON(t, recordStdout.Bytes(), &recordResult)
	if len(recordResult.Writes) != 2 {
		t.Fatalf("writes = %d, want 2", len(recordResult.Writes))
	}
	assertBloodPressureEntries(t, recordResult.Entries, []runner.BloodPressureEntry{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
	})

	list := `{"action":"list_blood_pressure","list_mode":"latest"}`
	var listStdout bytes.Buffer
	if err := run([]string{"blood-pressure", "--db", databasePath}, strings.NewReader(list), &listStdout, ioDiscard{}); err != nil {
		t.Fatalf("run blood-pressure list: %v", err)
	}
	var listResult runner.BloodPressureTaskResult
	decodeJSON(t, listStdout.Bytes(), &listResult)
	assertBloodPressureEntries(t, listResult.Entries, []runner.BloodPressureEntry{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
	})
}

func TestRunMedicationsJSONRoundTrip(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "nested", "openhealth.db")

	record := `{"action":"record_medications","medications":[{"name":"Levothyroxine","dosage_text":"25 mcg","start_date":"2026-01-01"},{"name":"Vitamin D","start_date":"2026-02-01","end_date":"2026-03-01"}]}`
	var recordStdout bytes.Buffer
	if err := run([]string{"medications", "--db", databasePath}, strings.NewReader(record), &recordStdout, ioDiscard{}); err != nil {
		t.Fatalf("run medications record: %v", err)
	}
	var recordResult runner.MedicationTaskResult
	decodeJSON(t, recordStdout.Bytes(), &recordResult)
	if len(recordResult.Writes) != 2 {
		t.Fatalf("writes = %d, want 2", len(recordResult.Writes))
	}
	assertMedicationEntries(t, recordResult.Entries, []runner.MedicationEntry{
		{Name: "Vitamin D", StartDate: "2026-02-01", EndDate: stringPointer("2026-03-01")},
		{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01"},
	})

	list := `{"action":"list_medications","status":"all"}`
	var listStdout bytes.Buffer
	if err := run([]string{"medications", "--db", databasePath}, strings.NewReader(list), &listStdout, ioDiscard{}); err != nil {
		t.Fatalf("run medications list: %v", err)
	}
	var listResult runner.MedicationTaskResult
	decodeJSON(t, listStdout.Bytes(), &listResult)
	assertMedicationEntries(t, listResult.Entries, []runner.MedicationEntry{
		{Name: "Vitamin D", StartDate: "2026-02-01", EndDate: stringPointer("2026-03-01")},
		{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01"},
	})
}

func TestRunLabsJSONRoundTrip(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "nested", "openhealth.db")

	record := `{"action":"record_labs","collections":[{"date":"2026-03-29","panels":[{"panel_name":"Metabolic","results":[{"test_name":"Glucose","canonical_slug":"glucose","value_text":"89","value_numeric":89,"units":"mg/dL","range_text":"70-99"}]}]},{"date":"2026-03-30","panels":[{"panel_name":"Thyroid","results":[{"test_name":"TSH","canonical_slug":"tsh","value_text":"3.1","units":"uIU/mL"}]}]}]}`
	var recordStdout bytes.Buffer
	if err := run([]string{"labs", "--db", databasePath}, strings.NewReader(record), &recordStdout, ioDiscard{}); err != nil {
		t.Fatalf("run labs record: %v", err)
	}
	var recordResult runner.LabTaskResult
	decodeJSON(t, recordStdout.Bytes(), &recordResult)
	if len(recordResult.Writes) != 2 {
		t.Fatalf("writes = %d, want 2", len(recordResult.Writes))
	}
	assertLabEntryDates(t, recordResult.Entries, []string{"2026-03-30", "2026-03-29"})

	list := `{"action":"list_labs","list_mode":"latest","analyte_slug":"glucose"}`
	var listStdout bytes.Buffer
	if err := run([]string{"labs", "--db", databasePath}, strings.NewReader(list), &listStdout, ioDiscard{}); err != nil {
		t.Fatalf("run labs list: %v", err)
	}
	var listResult runner.LabTaskResult
	decodeJSON(t, listStdout.Bytes(), &listResult)
	assertLabEntryDates(t, listResult.Entries, []string{"2026-03-29"})
}

func TestRunValidationRejectionDoesNotCreateDatabase(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "nested", "openhealth.db")
	request := `{"action":"upsert_weights","weights":[{"date":"2026-03-31","value":-5,"unit":"stone"}]}`

	var stdout bytes.Buffer
	if err := run([]string{"weight", "--db", databasePath}, strings.NewReader(request), &stdout, ioDiscard{}); err != nil {
		t.Fatalf("run invalid weight request: %v", err)
	}
	var result runner.WeightTaskResult
	decodeJSON(t, stdout.Bytes(), &result)
	if !result.Rejected || result.RejectionReason == "" {
		t.Fatalf("result = %#v, want rejection", result)
	}
	if _, err := os.Stat(databasePath); !os.IsNotExist(err) {
		t.Fatalf("database stat error = %v, want not exist", err)
	}
}

func TestRunRejectsBadJSONAndUnknownDomain(t *testing.T) {
	var stdout bytes.Buffer
	if err := run([]string{"weight"}, strings.NewReader("{"), &stdout, ioDiscard{}); err == nil || !strings.Contains(err.Error(), "decode request JSON") {
		t.Fatalf("bad JSON error = %v, want decode request JSON", err)
	}
	if err := run(nil, strings.NewReader("{}"), &stdout, ioDiscard{}); err == nil || !strings.Contains(err.Error(), "missing OpenHealth runner domain") {
		t.Fatalf("missing domain error = %v, want missing OpenHealth runner domain", err)
	}
	if err := run([]string{"unknown"}, strings.NewReader("{}"), &stdout, ioDiscard{}); err == nil || !strings.Contains(err.Error(), `unknown OpenHealth runner domain "unknown"`) {
		t.Fatalf("unknown domain error = %v, want unknown domain", err)
	}
}

func TestRunRejectsUnknownJSONFields(t *testing.T) {
	var stdout bytes.Buffer
	request := `{"action":"list_weights","list_mode":"latest","unexpected":true}`
	err := run([]string{"weight"}, strings.NewReader(request), &stdout, ioDiscard{})
	if err == nil || !strings.Contains(err.Error(), `unknown field "unexpected"`) {
		t.Fatalf("unknown field error = %v, want unknown field rejection", err)
	}
}

func TestRunReturnsRuntimeErrorsNonZeroForMain(t *testing.T) {
	databasePath := t.TempDir()
	request := `{"action":"list_weights","list_mode":"history","limit":1}`

	var stdout bytes.Buffer
	if err := run([]string{"weight", "--db", databasePath}, strings.NewReader(request), &stdout, ioDiscard{}); err == nil {
		t.Fatal("run weight with directory database path succeeded, want error")
	}
}

func TestRunDBFlagOverridesEnvironment(t *testing.T) {
	envDatabasePath := filepath.Join(t.TempDir(), "env", "openhealth.db")
	flagDatabasePath := filepath.Join(t.TempDir(), "flag", "openhealth.db")
	t.Setenv("OPENHEALTH_DATABASE_PATH", envDatabasePath)

	request := `{"action":"upsert_weights","weights":[{"date":"2026-03-29","value":152.2,"unit":"lb"}]}`
	var stdout bytes.Buffer
	if err := run([]string{"weight", "--db", flagDatabasePath}, strings.NewReader(request), &stdout, ioDiscard{}); err != nil {
		t.Fatalf("run with --db: %v", err)
	}
	if _, err := os.Stat(flagDatabasePath); err != nil {
		t.Fatalf("stat flag database path: %v", err)
	}
	if _, err := os.Stat(envDatabasePath); !os.IsNotExist(err) {
		t.Fatalf("env database stat error = %v, want not exist", err)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func decodeJSON[T any](t *testing.T, data []byte, out *T) {
	t.Helper()
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("decode JSON %q: %v", string(data), err)
	}
}

func assertWeightEntries(t *testing.T, got []runner.WeightTaskEntry, want []runner.WeightTaskEntry) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("entries = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("entries = %#v, want %#v", got, want)
		}
	}
}

func assertBloodPressureEntries(t *testing.T, got []runner.BloodPressureEntry, want []runner.BloodPressureEntry) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("entries = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i].Date != want[i].Date ||
			got[i].Systolic != want[i].Systolic ||
			got[i].Diastolic != want[i].Diastolic ||
			!equalIntPointers(got[i].Pulse, want[i].Pulse) {
			t.Fatalf("entries = %#v, want %#v", got, want)
		}
	}
}

func intPointer(value int) *int {
	return &value
}

func equalIntPointers(a *int, b *int) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}

func assertMedicationEntries(t *testing.T, got []runner.MedicationEntry, want []runner.MedicationEntry) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("entries = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i].Name != want[i].Name ||
			got[i].StartDate != want[i].StartDate ||
			!equalStringPointers(got[i].DosageText, want[i].DosageText) ||
			!equalStringPointers(got[i].EndDate, want[i].EndDate) {
			t.Fatalf("entries = %#v, want %#v", got, want)
		}
	}
}

func assertLabEntryDates(t *testing.T, got []runner.LabCollectionEntry, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("entries = %#v, want dates %#v", got, want)
	}
	for i := range want {
		if got[i].Date != want[i] {
			t.Fatalf("entries = %#v, want dates %#v", got, want)
		}
	}
}

func equalStringPointers(a *string, b *string) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}

func stringPointer(value string) *string {
	return &value
}
