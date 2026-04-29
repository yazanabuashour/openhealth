package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yazanabuashour/openhealth/client"
)

func TestMedicationScenarioExpandedCoverageVerifiesExpectedOutput(t *testing.T) {
	t.Parallel()

	sc, ok := scenarioByID("medication-non-oral-dosage")
	if !ok {
		t.Fatal("missing medication-non-oral-dosage scenario")
	}
	databasePath := filepath.Join(t.TempDir(), "openhealth.db")
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: databasePath})
	if err != nil {
		t.Fatalf("OpenLocal: %v", err)
	}
	if err := recordMedications(context.Background(), api, []medicationState{
		{Name: "Semaglutide", DosageText: stringPointer("2.5 mg subcutaneous injection weekly"), StartDate: "2026-02-01"},
	}); err != nil {
		t.Fatalf("recordMedications: %v", err)
	}
	_ = api.Close()

	verification, err := verifyScenario(databasePath, sc, "Semaglutide 2.5 mg subcutaneous injection weekly")
	if err != nil {
		t.Fatalf("verifyScenario: %v", err)
	}
	if !verification.Passed {
		t.Fatalf("verification failed: %#v", verification)
	}
	want := []medicationState{{Name: "Semaglutide", DosageText: stringPointer("2.5 mg subcutaneous injection weekly"), StartDate: "2026-02-01"}}
	if !medicationsEqual(verification.Medications, want) {
		t.Fatalf("medications = %s, want %s", describeMedications(verification.Medications), describeMedications(want))
	}
}

func TestMedicationNoteScenarioVerifiesExpectedOutput(t *testing.T) {
	t.Parallel()

	sc, ok := scenarioByID("medication-note")
	if !ok {
		t.Fatal("missing medication-note scenario")
	}
	databasePath := filepath.Join(t.TempDir(), "openhealth.db")
	api := openEvalTestClient(t, databasePath)
	if err := recordMedications(context.Background(), api, []medicationState{
		{Name: "Semaglutide", DosageText: stringPointer("0.25 mg subcutaneous injection weekly"), StartDate: "2026-02-01", Note: stringPointer("coverage approved after prior authorization")},
	}); err != nil {
		t.Fatalf("recordMedications: %v", err)
	}
	_ = api.Close()

	verification, err := verifyScenario(databasePath, sc, "Semaglutide subcutaneous coverage approved")
	if err != nil {
		t.Fatalf("verifyScenario: %v", err)
	}
	if !verification.Passed {
		t.Fatalf("verification failed: %#v", verification)
	}
	want := []medicationState{{Name: "Semaglutide", DosageText: stringPointer("0.25 mg subcutaneous injection weekly"), StartDate: "2026-02-01", Note: stringPointer("coverage approved after prior authorization")}}
	if !medicationsEqual(verification.Medications, want) {
		t.Fatalf("medications = %s, want %s", describeMedications(verification.Medications), describeMedications(want))
	}
}

func TestLabScenarioExpandedCoverageVerifiesExpectedOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		scenarioID   string
		finalMessage string
		seed         func(t *testing.T, databasePath string)
		wantLabs     []labCollectionState
	}{
		{
			scenarioID:   "lab-arbitrary-slug",
			finalMessage: "2026-03-29 Vitamin D 32 ng/mL",
			seed: func(t *testing.T, databasePath string) {
				t.Helper()
				api := openEvalTestClient(t, databasePath)
				defer func() { _ = api.Close() }()
				if err := recordLabs(context.Background(), api, []labCollectionState{
					{Date: "2026-03-29", Results: []labResultState{{TestName: "Vitamin D", CanonicalSlug: stringPointer("vitamin-d"), ValueText: "32", ValueNumeric: floatPointer(32), Units: stringPointer("ng/mL")}}},
				}); err != nil {
					t.Fatalf("recordLabs: %v", err)
				}
			},
			wantLabs: []labCollectionState{{Date: "2026-03-29", Results: []labResultState{{TestName: "Vitamin D", CanonicalSlug: stringPointer("vitamin-d"), ValueText: "32", ValueNumeric: floatPointer(32), Units: stringPointer("ng/mL")}}}},
		},
		{
			scenarioID:   "lab-note",
			finalMessage: "2026-03-29 Glucose 89 mg/dL; labs look stable; A1C context",
			seed: func(t *testing.T, databasePath string) {
				t.Helper()
				api := openEvalTestClient(t, databasePath)
				defer func() { _ = api.Close() }()
				if err := recordLabs(context.Background(), api, []labCollectionState{
					{Date: "2026-03-29", Note: stringPointer("labs look stable, keep moving"), Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL"), Notes: []string{"HIV 4th gen narrative", "A1C context"}}}},
				}); err != nil {
					t.Fatalf("recordLabs: %v", err)
				}
			},
			wantLabs: []labCollectionState{{Date: "2026-03-29", Note: stringPointer("labs look stable, keep moving"), Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL"), Notes: []string{"HIV 4th gen narrative", "A1C context"}}}}},
		},
		{
			scenarioID:   "lab-same-day-multiple",
			finalMessage: "2026-03-29 TSH 3.1 uIU/mL and 2026-03-29 Glucose 89 mg/dL",
			seed: func(t *testing.T, databasePath string) {
				t.Helper()
				api := openEvalTestClient(t, databasePath)
				defer func() { _ = api.Close() }()
				if err := recordLabs(context.Background(), api, []labCollectionState{
					{Date: "2026-03-29", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")}}},
					{Date: "2026-03-29", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.1", ValueNumeric: floatPointer(3.1), Units: stringPointer("uIU/mL")}}},
				}); err != nil {
					t.Fatalf("recordLabs: %v", err)
				}
			},
			wantLabs: []labCollectionState{
				{Date: "2026-03-29", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.1", ValueNumeric: floatPointer(3.1), Units: stringPointer("uIU/mL")}}},
				{Date: "2026-03-29", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")}}},
			},
		},
		{
			scenarioID:   "lab-patch",
			finalMessage: "2026-03-29 Glucose 92 mg/dL; HDL 51 mg/dL",
			seed: func(t *testing.T, databasePath string) {
				t.Helper()
				sc, ok := scenarioByID("lab-patch")
				if !ok {
					t.Fatal("missing lab-patch scenario")
				}
				if err := seedScenario(databasePath, sc); err != nil {
					t.Fatalf("seedScenario: %v", err)
				}
				api := openEvalTestClient(t, databasePath)
				defer func() { _ = api.Close() }()
				collections, err := api.ListLabCollections(context.Background())
				if err != nil {
					t.Fatalf("ListLabCollections: %v", err)
				}
				if len(collections) != 1 {
					t.Fatalf("collections = %#v, want one", collections)
				}
				if _, err := api.ReplaceLabCollection(context.Background(), collections[0].ID, client.LabCollectionInput{
					CollectedAt: collections[0].CollectedAt,
					Panels: []client.LabPanelInput{{PanelName: "Panel", Results: []client.LabResultInput{
						{TestName: "Glucose", CanonicalSlug: clientAnalyteSlug(stringPointer("glucose")), ValueText: "92", ValueNumeric: floatPointer(92), Units: stringPointer("mg/dL")},
						{TestName: "HDL", CanonicalSlug: clientAnalyteSlug(stringPointer("hdl")), ValueText: "51", ValueNumeric: floatPointer(51), Units: stringPointer("mg/dL")},
					}}},
				}); err != nil {
					t.Fatalf("ReplaceLabCollection: %v", err)
				}
			},
			wantLabs: []labCollectionState{{Date: "2026-03-29", Results: []labResultState{
				{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "92", ValueNumeric: floatPointer(92), Units: stringPointer("mg/dL")},
				{TestName: "HDL", CanonicalSlug: stringPointer("hdl"), ValueText: "51", ValueNumeric: floatPointer(51), Units: stringPointer("mg/dL")},
			}}},
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
			tt.seed(t, databasePath)

			verification, err := verifyScenario(databasePath, sc, tt.finalMessage)
			if err != nil {
				t.Fatalf("verifyScenario: %v", err)
			}
			if !verification.Passed {
				t.Fatalf("verification failed: %#v", verification)
			}
			if !labsEqual(verification.Labs, tt.wantLabs) {
				t.Fatalf("labs = %s, want %s", describeLabs(verification.Labs), describeLabs(tt.wantLabs))
			}
		})
	}
}

func TestBodyCompositionCombinedRowScenarioVerifiesExpectedOutput(t *testing.T) {
	t.Parallel()

	sc, ok := scenarioByID("body-composition-combined-weight-row")
	if !ok {
		t.Fatal("missing body-composition-combined-weight-row scenario")
	}
	databasePath := filepath.Join(t.TempDir(), "openhealth.db")
	api := openEvalTestClient(t, databasePath)
	if err := upsertWeights(context.Background(), api, []weightState{{Date: "2026-03-29", Value: 154.2, Unit: "lb", Note: stringPointer("smart scale")}}); err != nil {
		t.Fatalf("upsertWeights: %v", err)
	}
	if err := recordBodyComposition(context.Background(), api, []bodyCompositionState{
		{Date: "2026-03-29", BodyFatPercent: floatPointer(18.7), WeightValue: floatPointer(154.2), WeightUnit: stringPointer("lb"), Method: stringPointer("smart scale"), Note: stringPointer("smart scale")},
	}); err != nil {
		t.Fatalf("recordBodyComposition: %v", err)
	}
	_ = api.Close()

	verification, err := verifyScenario(databasePath, sc, "Stored weight 154.2 lb and body fat 18.7%.")
	if err != nil {
		t.Fatalf("verifyScenario: %v", err)
	}
	if !verification.Passed {
		t.Fatalf("verification failed: %#v", verification)
	}
	wantWeights := []weightState{{Date: "2026-03-29", Value: 154.2, Unit: "lb"}}
	wantBody := []bodyCompositionState{{Date: "2026-03-29", BodyFatPercent: floatPointer(18.7), WeightValue: floatPointer(154.2), WeightUnit: stringPointer("lb"), Method: stringPointer("smart scale")}}
	if !weightsEqual(verification.Weights, wantWeights) {
		t.Fatalf("weights = %s, want %s", describeWeights(verification.Weights), describeWeights(wantWeights))
	}
	if !bodyCompositionEqual(verification.BodyComposition, wantBody) {
		t.Fatalf("body composition = %s, want %s", describeBodyComposition(verification.BodyComposition), describeBodyComposition(wantBody))
	}
}

func TestImagingScenariosVerifyExpectedOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		scenarioID   string
		finalMessage string
		seed         func(t *testing.T, databasePath string)
		wantImaging  []imagingState
	}{
		{
			scenarioID:   "imaging-record-list",
			finalMessage: "2026-03-29 chest X-ray narrative",
			seed: func(t *testing.T, databasePath string) {
				t.Helper()
				api := openEvalTestClient(t, databasePath)
				defer func() { _ = api.Close() }()
				if err := recordImaging(context.Background(), api, []imagingState{
					{Date: "2026-03-29", Modality: "X-ray", BodySite: stringPointer("chest"), Title: stringPointer("Chest X-ray"), Summary: "No acute cardiopulmonary abnormality.", Impression: stringPointer("Normal chest radiograph."), Note: stringPointer("ordered for cough"), Notes: []string{"XR TOE RIGHT narrative", "US Head/Neck findings"}},
				}); err != nil {
					t.Fatalf("recordImaging: %v", err)
				}
			},
			wantImaging: []imagingState{{Date: "2026-03-29", Modality: "X-ray", BodySite: stringPointer("chest"), Title: stringPointer("Chest X-ray"), Summary: "No acute cardiopulmonary abnormality", Impression: stringPointer("Normal chest radiograph"), Note: stringPointer("ordered for cough"), Notes: []string{"XR TOE RIGHT narrative", "US Head/Neck findings"}}},
		},
		{
			scenarioID:   "imaging-correct",
			finalMessage: "2026-03-29 CT Stable small pulmonary nodule.",
			seed: func(t *testing.T, databasePath string) {
				t.Helper()
				if err := seedScenario(databasePath, scenario{ID: "imaging-correct"}); err != nil {
					t.Fatalf("seedScenario: %v", err)
				}
				api := openEvalTestClient(t, databasePath)
				defer func() { _ = api.Close() }()
				records, err := api.ListImaging(context.Background(), client.ImagingListOptions{})
				if err != nil {
					t.Fatalf("ListImaging: %v", err)
				}
				if len(records) != 1 {
					t.Fatalf("records = %#v, want one", records)
				}
				if _, err := api.ReplaceImaging(context.Background(), records[0].ID, client.ImagingRecordInput{
					PerformedAt: records[0].PerformedAt,
					Modality:    "CT",
					BodySite:    stringPointer("chest"),
					Title:       stringPointer("Chest X-ray"),
					Summary:     "Stable small pulmonary nodule.",
					Impression:  stringPointer("Normal chest radiograph."),
					Note:        stringPointer("ordered for cough"),
				}); err != nil {
					t.Fatalf("ReplaceImaging: %v", err)
				}
			},
			wantImaging: []imagingState{{Date: "2026-03-29", Modality: "CT", BodySite: stringPointer("chest"), Title: stringPointer("Chest X-ray"), Summary: "Stable small pulmonary nodule.", Impression: stringPointer("Normal chest radiograph."), Note: stringPointer("ordered for cough")}},
		},
		{
			scenarioID:   "imaging-delete",
			finalMessage: "Deleted imaging record.",
			seed: func(t *testing.T, databasePath string) {
				t.Helper()
				if err := seedScenario(databasePath, scenario{ID: "imaging-delete"}); err != nil {
					t.Fatalf("seedScenario: %v", err)
				}
				api := openEvalTestClient(t, databasePath)
				defer func() { _ = api.Close() }()
				records, err := api.ListImaging(context.Background(), client.ImagingListOptions{})
				if err != nil {
					t.Fatalf("ListImaging: %v", err)
				}
				if len(records) != 1 {
					t.Fatalf("records = %#v, want one", records)
				}
				if err := api.DeleteImaging(context.Background(), records[0].ID); err != nil {
					t.Fatalf("DeleteImaging: %v", err)
				}
			},
			wantImaging: []imagingState{},
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
			tt.seed(t, databasePath)
			verification, err := verifyScenario(databasePath, sc, tt.finalMessage)
			if err != nil {
				t.Fatalf("verifyScenario: %v", err)
			}
			if !verification.Passed {
				t.Fatalf("verification failed: %#v", verification)
			}
			if !imagingEqual(verification.Imaging, tt.wantImaging) {
				t.Fatalf("imaging = %s, want %s", describeImaging(verification.Imaging), describeImaging(tt.wantImaging))
			}
		})
	}
}

func TestMixedImportFileCoverageScenarioVerifiesExpectedOutput(t *testing.T) {
	t.Parallel()

	sc, ok := scenarioByID("mixed-import-file-coverage")
	if !ok {
		t.Fatal("missing mixed-import-file-coverage scenario")
	}
	databasePath := filepath.Join(t.TempDir(), "openhealth.db")
	api := openEvalTestClient(t, databasePath)
	if err := upsertWeights(context.Background(), api, []weightState{{Date: "2026-03-29", Value: 154.2, Unit: "lb", Note: stringPointer("morning scale")}}); err != nil {
		t.Fatalf("upsertWeights: %v", err)
	}
	if err := recordBodyComposition(context.Background(), api, []bodyCompositionState{{Date: "2026-03-29", BodyFatPercent: floatPointer(18.7), WeightValue: floatPointer(154.2), WeightUnit: stringPointer("lb")}}); err != nil {
		t.Fatalf("recordBodyComposition: %v", err)
	}
	if err := recordLabs(context.Background(), api, []labCollectionState{{Date: "2026-03-29", Note: stringPointer("labs look stable, keep moving"), Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL"), Notes: []string{"HIV 4th gen narrative", "A1C context"}}}}}); err != nil {
		t.Fatalf("recordLabs: %v", err)
	}
	if err := recordImaging(context.Background(), api, []imagingState{{Date: "2026-03-29", Modality: "X-ray", BodySite: stringPointer("chest"), Title: stringPointer("Chest X-ray"), Summary: "No acute cardiopulmonary abnormality.", Impression: stringPointer("Normal chest radiograph."), Notes: []string{"XR TOE RIGHT narrative", "US Head/Neck findings"}}}); err != nil {
		t.Fatalf("recordImaging: %v", err)
	}
	if err := recordMedications(context.Background(), api, []medicationState{{Name: "Semaglutide", DosageText: stringPointer("0.25 mg subcutaneous injection weekly"), StartDate: "2026-02-01", Note: stringPointer("coverage approved after prior authorization")}}); err != nil {
		t.Fatalf("recordMedications: %v", err)
	}
	_ = api.Close()

	verification, err := verifyScenario(databasePath, sc, "154.2 18.7 Glucose 89 Semaglutide X-ray narrative")
	if err != nil {
		t.Fatalf("verifyScenario: %v", err)
	}
	if !verification.Passed {
		t.Fatalf("verification failed: %#v", verification)
	}
}

func TestMixedAndMultiTurnScenariosVerifyExpectedOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		scenarioID   string
		turnIndex    int
		finalMessage string
		wantWeights  []weightState
		wantReadings []bloodPressureState
		manualSeed   bool
	}{
		{
			scenarioID:   "mixed-add-latest",
			turnIndex:    1,
			finalMessage: "Recorded on 2026-03-31.\n\nLatest weight: 150.8 lb\nLatest blood pressure: 119/77, pulse 62",
			wantWeights:  []weightState{{Date: "2026-03-31", Value: 150.8, Unit: "lb"}},
			wantReadings: []bloodPressureState{{Date: "2026-03-31", Systolic: 119, Diastolic: 77, Pulse: intPointer(62)}},
			manualSeed:   true,
		},
		{
			scenarioID: "mixed-bounded-range",
			turnIndex:  1,
			finalMessage: strings.Join([]string{
				"Weight 2026-03-30 151.6 lb",
				"Weight 2026-03-29 152.2 lb",
				"Blood pressure 2026-03-30 118/76",
				"Blood pressure 2026-03-29 122/78 pulse 64",
			}, "\n"),
			wantWeights: []weightState{
				{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
				{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
				{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
			},
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
				{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			},
		},
		{
			scenarioID:   "mixed-invalid-direct-reject",
			turnIndex:    1,
			finalMessage: "Invalid request: weight unit stone is unsupported and blood-pressure values must be positive.",
			wantWeights:  []weightState{},
			wantReadings: []bloodPressureState{},
		},
		{
			scenarioID:   "mt-weight-clarify-then-add",
			turnIndex:    1,
			finalMessage: "Which year should I use for 03/29?",
			wantWeights:  []weightState{},
			wantReadings: []bloodPressureState{},
		},
		{
			scenarioID:   "mt-mixed-latest-then-correct",
			turnIndex:    2,
			finalMessage: "Updated 2026-03-30: weight 151.0 lb and blood pressure 117/75 pulse 63.",
			wantWeights: []weightState{
				{Date: "2026-03-30", Value: 151.0, Unit: "lb"},
				{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			},
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 117, Diastolic: 75, Pulse: intPointer(63)},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			},
			manualSeed: true,
		},
		{
			scenarioID:   "mt-bp-latest-then-correct",
			turnIndex:    1,
			finalMessage: "2026-03-30 118/76",
			wantWeights:  []weightState{},
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			},
		},
		{
			scenarioID:   "mt-bp-latest-then-correct",
			turnIndex:    2,
			finalMessage: `{"date":"2026-03-30","systolic":117,"diastolic":75,"pulse":63}`,
			wantWeights:  []weightState{},
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 117, Diastolic: 75, Pulse: intPointer(63)},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			},
			manualSeed: true,
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
			if !tt.manualSeed {
				if err := seedScenario(databasePath, sc); err != nil {
					t.Fatalf("seedScenario: %v", err)
				}
			}
			if tt.manualSeed {
				api, err := client.OpenLocal(client.LocalConfig{DatabasePath: databasePath})
				if err != nil {
					t.Fatalf("OpenLocal: %v", err)
				}
				if err := upsertWeights(context.Background(), api, tt.wantWeights); err != nil {
					t.Fatalf("upsertWeights: %v", err)
				}
				if err := recordBloodPressures(context.Background(), api, tt.wantReadings); err != nil {
					t.Fatalf("recordBloodPressures: %v", err)
				}
				_ = api.Close()
			}

			verification, err := verifyScenarioTurn(databasePath, sc, tt.turnIndex, tt.finalMessage)
			if err != nil {
				t.Fatalf("verifyScenarioTurn: %v", err)
			}
			if !verification.Passed {
				t.Fatalf("verification failed: %#v", verification)
			}
			if !weightsEqual(verification.Weights, tt.wantWeights) {
				t.Fatalf("weights = %s, want %s", describeWeights(verification.Weights), describeWeights(tt.wantWeights))
			}
			if !bloodPressuresEqual(verification.BloodPressures, tt.wantReadings) {
				t.Fatalf("blood pressures = %s, want %s", describeBloodPressures(verification.BloodPressures), describeBloodPressures(tt.wantReadings))
			}
		})
	}
}
