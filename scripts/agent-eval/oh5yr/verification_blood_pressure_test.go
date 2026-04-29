package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yazanabuashour/openhealth/client"
)

func TestBloodPressureScenariosVerifyExpectedOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		scenarioID   string
		finalMessage string
		wantReadings []bloodPressureState
	}{
		{
			scenarioID: "bp-latest-only",
			finalMessage: strings.Join([]string{
				"2026-03-30 118/76",
			}, "\n"),
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
				{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			},
		},
		{
			scenarioID: "bp-history-limit-two",
			finalMessage: strings.Join([]string{
				"2026-03-30 118/76",
				"2026-03-29 122/78 pulse 64",
			}, "\n"),
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
				{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
				{Date: "2026-03-27", Systolic: 126, Diastolic: 82},
			},
		},
		{
			scenarioID: "bp-bounded-range",
			finalMessage: strings.Join([]string{
				"2026-03-30 118/76",
				"2026-03-29 122/78 pulse 64",
			}, "\n"),
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
				{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			},
		},
		{
			scenarioID: "bp-bounded-range-natural",
			finalMessage: strings.Join([]string{
				"March 30, 2026",
				"- 118/76",
				"",
				"March 29, 2026",
				"- 122/78, pulse 64",
			}, "\n"),
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
				{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			},
		},
		{
			scenarioID:   "bp-invalid-input",
			finalMessage: "Invalid blood pressure: systolic, diastolic, and pulse must be positive.",
			wantReadings: []bloodPressureState{},
		},
		{
			scenarioID:   "bp-invalid-relation",
			finalMessage: "Invalid blood pressure: systolic must be greater than diastolic.",
			wantReadings: []bloodPressureState{},
		},
		{
			scenarioID:   "bp-non-iso-date-reject",
			finalMessage: "Invalid date: use YYYY-MM-DD.",
			wantReadings: []bloodPressureState{},
		},
		{
			scenarioID:   "bp-correct-existing",
			finalMessage: "Updated 2026-03-29 to 121/77 pulse 63.",
			wantReadings: []bloodPressureState{
				{Date: "2026-03-29", Systolic: 121, Diastolic: 77, Pulse: intPointer(63)},
			},
		},
		{
			scenarioID:   "bp-correct-missing-reject",
			finalMessage: "No update was made for 2026-03-31 because there is no local blood-pressure reading on that date to correct.",
			wantReadings: []bloodPressureState{
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			},
		},
		{
			scenarioID:   "bp-correct-ambiguous-reject",
			finalMessage: "Multiple readings exist for 2026-03-29, so the correction is ambiguous and was not updated.",
			wantReadings: []bloodPressureState{
				{Date: "2026-03-29", Systolic: 120, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.scenarioID, func(t *testing.T) {
			t.Parallel()
			sc, ok := scenarioByID(tt.scenarioID)
			if !ok {
				t.Fatalf("missing scenario %q", tt.scenarioID)
			}
			databasePath := filepath.Join(t.TempDir(), "openhealth.db")
			if err := seedScenario(databasePath, sc); err != nil {
				t.Fatalf("seedScenario: %v", err)
			}
			if tt.scenarioID == "bp-correct-existing" {
				api, err := client.OpenLocal(client.LocalConfig{DatabasePath: databasePath})
				if err != nil {
					t.Fatalf("OpenLocal: %v", err)
				}
				readings, err := api.ListBloodPressure(context.Background(), client.BloodPressureListOptions{Limit: 1})
				if err != nil {
					t.Fatalf("ListBloodPressure: %v", err)
				}
				if len(readings) != 1 {
					t.Fatalf("seed readings = %d, want 1", len(readings))
				}
				pulse63 := 63
				if _, err := api.ReplaceBloodPressure(context.Background(), readings[0].ID, client.BloodPressureRecordInput{
					RecordedAt: readings[0].RecordedAt,
					Systolic:   121,
					Diastolic:  77,
					Pulse:      &pulse63,
				}); err != nil {
					t.Fatalf("ReplaceBloodPressure: %v", err)
				}
				_ = api.Close()
			}
			verification, err := verifyScenario(databasePath, sc, tt.finalMessage)
			if err != nil {
				t.Fatalf("verifyScenario: %v", err)
			}
			if !verification.Passed {
				t.Fatalf("verification failed: %#v", verification)
			}
			if !bloodPressuresEqual(verification.BloodPressures, tt.wantReadings) {
				t.Fatalf("blood pressures = %s, want %s", describeBloodPressures(verification.BloodPressures), describeBloodPressures(tt.wantReadings))
			}
		})
	}
}

func TestBloodPressureNoteScenarioVerifiesExpectedOutput(t *testing.T) {
	t.Parallel()

	sc, ok := scenarioByID("bp-add-two")
	if !ok {
		t.Fatal("missing bp-add-two scenario")
	}
	databasePath := filepath.Join(t.TempDir(), "openhealth.db")
	api := openEvalTestClient(t, databasePath)
	if err := recordBloodPressures(context.Background(), api, []bloodPressureState{
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64), Note: stringPointer("home cuff")},
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
	}); err != nil {
		t.Fatalf("recordBloodPressures: %v", err)
	}
	_ = api.Close()

	verification, err := verifyScenario(databasePath, sc, "Stored 2026-03-29 122/78 pulse 64 home cuff and 2026-03-30 118/76.")
	if err != nil {
		t.Fatalf("verifyScenario: %v", err)
	}
	if !verification.Passed {
		t.Fatalf("verification failed: %#v", verification)
	}
	want := []bloodPressureState{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64), Note: stringPointer("home cuff")},
	}
	if !bloodPressuresEqual(verification.BloodPressures, want) {
		t.Fatalf("blood pressures = %s, want %s", describeBloodPressures(verification.BloodPressures), describeBloodPressures(want))
	}
}
