package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yazanabuashour/openhealth/client"
)

func TestExpandedWeightScenariosVerifyExpectedOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		scenarioID   string
		finalMessage string
		wantWeights  []weightState
	}{
		{
			name:         "latest text",
			scenarioID:   "latest-only",
			finalMessage: "2026-03-30 151.6 lb",
			wantWeights: []weightState{
				{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
				{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
				{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
			},
		},
		{
			name:       "history text",
			scenarioID: "history-limit-two",
			finalMessage: strings.Join([]string{
				"2026-03-30 151.6 lb",
				"2026-03-29 152.2 lb",
			}, "\n"),
			wantWeights: []weightState{
				{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
				{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
				{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
				{Date: "2026-03-27", Value: 154.1, Unit: "lb"},
			},
		},
		{
			name:       "history json lines",
			scenarioID: "history-limit-two",
			finalMessage: strings.Join([]string{
				`{"date":"2026-03-30","value":151.6,"unit":"lb"}`,
				`{"date":"2026-03-29","value":152.2,"unit":"lb"}`,
			}, "\n"),
			wantWeights: []weightState{
				{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
				{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
				{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
				{Date: "2026-03-27", Value: 154.1, Unit: "lb"},
			},
		},
		{
			name:         "non iso reject",
			scenarioID:   "non-iso-date-reject",
			finalMessage: "Invalid date: use YYYY-MM-DD.",
			wantWeights:  []weightState{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sc, ok := scenarioByID(tt.scenarioID)
			if !ok {
				t.Fatalf("missing scenario %q", tt.scenarioID)
			}
			databasePath := filepath.Join(t.TempDir(), "openhealth.db")
			if err := seedScenario(databasePath, sc); err != nil {
				t.Fatalf("seedScenario: %v", err)
			}
			verification, err := verifyScenario(databasePath, sc, tt.finalMessage)
			if err != nil {
				t.Fatalf("verifyScenario: %v", err)
			}
			if !verification.Passed {
				t.Fatalf("verification failed: %#v", verification)
			}
			if !weightsEqual(verification.Weights, tt.wantWeights) {
				t.Fatalf("weights = %s, want %s", describeWeights(verification.Weights), describeWeights(tt.wantWeights))
			}
		})
	}
}

func TestWeightOnlyMultiTurnRejectsBloodPressureWrites(t *testing.T) {
	t.Parallel()

	sc, ok := scenarioByID("mt-weight-clarify-then-add")
	if !ok {
		t.Fatal("missing mt-weight-clarify-then-add scenario")
	}
	tests := []struct {
		name         string
		turnIndex    int
		finalMessage string
		weights      []weightState
		readings     []bloodPressureState
	}{
		{
			name:         "turn one",
			turnIndex:    1,
			finalMessage: "Which year should I use for 03/29?",
			readings:     []bloodPressureState{{Date: "2026-03-29", Systolic: 122, Diastolic: 78}},
		},
		{
			name:         "turn two",
			turnIndex:    2,
			finalMessage: "Stored 2026-03-29 152.2 lb.",
			weights:      []weightState{{Date: "2026-03-29", Value: 152.2, Unit: "lb"}},
			readings:     []bloodPressureState{{Date: "2026-03-29", Systolic: 122, Diastolic: 78}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			databasePath := filepath.Join(t.TempDir(), "openhealth.db")
			api, err := client.OpenLocal(client.LocalConfig{DatabasePath: databasePath})
			if err != nil {
				t.Fatalf("OpenLocal: %v", err)
			}
			if len(tt.weights) > 0 {
				if err := upsertWeights(context.Background(), api, tt.weights); err != nil {
					t.Fatalf("upsertWeights: %v", err)
				}
			}
			if err := recordBloodPressures(context.Background(), api, tt.readings); err != nil {
				t.Fatalf("recordBloodPressures: %v", err)
			}
			_ = api.Close()

			verification, err := verifyScenarioTurn(databasePath, sc, tt.turnIndex, tt.finalMessage)
			if err != nil {
				t.Fatalf("verifyScenarioTurn: %v", err)
			}
			if verification.DatabasePass || verification.Passed {
				t.Fatalf("verification = %#v, want database failure for stray blood-pressure write", verification)
			}
		})
	}
}

func TestNaturalBoundedRangeScenarioSeedsExpectedRows(t *testing.T) {
	t.Parallel()

	sc, ok := scenarioByID("bounded-range-natural")
	if !ok {
		t.Fatal("missing bounded-range-natural scenario")
	}
	if !strings.Contains(sc.Prompt, "Mar 29") || !strings.Contains(sc.Prompt, "Mar 30") {
		t.Fatalf("prompt = %q, want natural Mar 29 and Mar 30 wording", sc.Prompt)
	}

	databasePath := filepath.Join(t.TempDir(), "openhealth.db")
	if err := seedScenario(databasePath, sc); err != nil {
		t.Fatalf("seedScenario: %v", err)
	}
	weights, err := listWeights(databasePath)
	if err != nil {
		t.Fatalf("listWeights: %v", err)
	}
	got := weightStates(weights)
	want := []weightState{
		{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
		{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
	}
	if !weightsEqual(got, want) {
		t.Fatalf("seeded weights = %s, want %s", describeWeights(got), describeWeights(want))
	}
}
