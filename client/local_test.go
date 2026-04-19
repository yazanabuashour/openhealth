package client_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/yazanabuashour/openhealth/client"
)

func TestOpenLocalSupportsDirectLocalClient(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	api, err := client.OpenLocal(client.LocalConfig{
		DataDir: dataDir,
	})
	if err != nil {
		t.Fatalf("open local client: %v", err)
	}
	defer func() {
		if closeErr := api.Close(); closeErr != nil {
			t.Fatalf("close local client: %v", closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	summary, err := api.Summary(ctx)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.ActiveMedicationCount != 0 {
		t.Fatalf("activeMedicationCount = %d, want 0", summary.ActiveMedicationCount)
	}

	recordedAt := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	created, err := api.RecordWeight(ctx, client.WeightRecordInput{
		RecordedAt: recordedAt,
		Unit:       client.WeightUnitLb,
		Value:      149.4,
	})
	if err != nil {
		t.Fatalf("record weight: %v", err)
	}
	if created.Value != 149.4 {
		t.Fatalf("created value = %v, want 149.4", created.Value)
	}

	weights, err := api.ListWeights(ctx, client.WeightListOptions{})
	if err != nil {
		t.Fatalf("list weights: %v", err)
	}
	if len(weights) != 1 || weights[0].Value != 149.4 {
		t.Fatalf("weights = %#v, want one 149.4 entry", weights)
	}

	databasePath := filepath.Join(dataDir, "openhealth.db")
	if api.Paths.DatabasePath != databasePath {
		t.Fatalf("databasePath = %q, want %q", api.Paths.DatabasePath, databasePath)
	}
	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("stat database path: %v", err)
	}
}

func TestOpenLocalPersistsDataAcrossSessions(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	recordedAt := time.Date(2026, 4, 14, 12, 0, 0, 0, time.UTC)

	api, err := client.OpenLocal(client.LocalConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("open first local client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	created, err := api.RecordWeight(ctx, client.WeightRecordInput{
		RecordedAt: recordedAt,
		Unit:       client.WeightUnitLb,
		Value:      152.1,
	})
	if err != nil {
		t.Fatalf("record first session weight: %v", err)
	}
	if created.Value != 152.1 {
		t.Fatalf("created value = %v, want 152.1", created.Value)
	}
	if err := api.Close(); err != nil {
		t.Fatalf("close first local client: %v", err)
	}

	api, err = client.OpenLocal(client.LocalConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("reopen local client: %v", err)
	}
	defer func() {
		if closeErr := api.Close(); closeErr != nil {
			t.Fatalf("close reopened local client: %v", closeErr)
		}
	}()

	weights, err := api.ListWeights(ctx, client.WeightListOptions{})
	if err != nil {
		t.Fatalf("list second session weights: %v", err)
	}
	if len(weights) != 1 || weights[0].Value != 152.1 {
		t.Fatalf("weights after reopen = %#v, want one 152.1 entry", weights)
	}
}

func TestLocalClientWeightHelpers(t *testing.T) {
	t.Parallel()

	api, err := client.OpenLocal(client.LocalConfig{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("open local client: %v", err)
	}
	defer func() {
		if closeErr := api.Close(); closeErr != nil {
			t.Fatalf("close local client: %v", closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	march29 := time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)
	created, err := api.UpsertWeight(ctx, client.WeightRecordInput{
		RecordedAt: march29,
		Value:      152.2,
		Unit:       client.WeightUnitLb,
		Note:       stringPointer("morning scale"),
	})
	if err != nil {
		t.Fatalf("upsert weight: %v", err)
	}
	if created.Status != client.WeightWriteStatusCreated || created.Entry.Value != 152.2 {
		t.Fatalf("created weight = %#v", created)
	}
	if created.Entry.Note == nil || *created.Entry.Note != "morning scale" {
		t.Fatalf("created weight note = %#v", created.Entry.Note)
	}

	again, err := api.UpsertWeight(ctx, client.WeightRecordInput{
		RecordedAt: march29,
		Value:      152.2,
		Unit:       client.WeightUnitLb,
	})
	if err != nil {
		t.Fatalf("repeat upsert weight: %v", err)
	}
	if again.Status != client.WeightWriteStatusAlreadyExists || again.Entry.ID != created.Entry.ID {
		t.Fatalf("repeat weight = %#v, want already_exists id %d", again, created.Entry.ID)
	}

	march30 := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)
	recorded, err := api.RecordWeight(ctx, client.WeightRecordInput{
		RecordedAt: march30,
		Value:      151.6,
	})
	if err != nil {
		t.Fatalf("record weight with default unit: %v", err)
	}
	if recorded.Unit != client.WeightUnitLb {
		t.Fatalf("recorded unit = %q, want %q", recorded.Unit, client.WeightUnitLb)
	}

	weights, err := api.ListWeights(ctx, client.WeightListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("list weights: %v", err)
	}
	if len(weights) != 2 {
		t.Fatalf("weight count = %d, want 2", len(weights))
	}
	if weights[0].Value != 151.6 || !weights[0].RecordedAt.Equal(march30) {
		t.Fatalf("newest weight = %#v, want 151.6 on March 30", weights[0])
	}

	latest, err := api.LatestWeight(ctx)
	if err != nil {
		t.Fatalf("latest weight: %v", err)
	}
	if latest == nil || latest.ID != weights[0].ID {
		t.Fatalf("latest = %#v, want id %d", latest, weights[0].ID)
	}
}

func TestLocalClientNonWeightHelpers(t *testing.T) {
	t.Parallel()

	api, err := client.OpenLocal(client.LocalConfig{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("open local client: %v", err)
	}
	defer func() {
		if closeErr := api.Close(); closeErr != nil {
			t.Fatalf("close local client: %v", closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	recordedAt := time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC)
	bp, err := api.RecordBloodPressure(ctx, client.BloodPressureRecordInput{
		RecordedAt: recordedAt,
		Systolic:   118,
		Diastolic:  76,
		Pulse:      intPointer(64),
		Note:       stringPointer("home cuff"),
	})
	if err != nil {
		t.Fatalf("record blood pressure: %v", err)
	}
	bp, err = api.ReplaceBloodPressure(ctx, bp.ID, client.BloodPressureRecordInput{
		RecordedAt: recordedAt.Add(time.Hour),
		Systolic:   119,
		Diastolic:  77,
		Note:       stringPointer("manual correction"),
	})
	if err != nil {
		t.Fatalf("replace blood pressure: %v", err)
	}
	if bp.Systolic != 119 || bp.Pulse != nil {
		t.Fatalf("blood pressure = %#v, want systolic 119 and nil pulse", bp)
	}
	if bp.Note == nil || *bp.Note != "manual correction" {
		t.Fatalf("blood pressure note = %#v", bp.Note)
	}
	bps, err := api.ListBloodPressure(ctx, client.BloodPressureListOptions{Limit: 1})
	if err != nil {
		t.Fatalf("list blood pressure: %v", err)
	}
	if len(bps) != 1 || bps[0].ID != bp.ID {
		t.Fatalf("blood pressure list = %#v", bps)
	}

	med, err := api.CreateMedicationCourse(ctx, client.MedicationCourseInput{
		Name:       "Levothyroxine",
		DosageText: stringPointer("25 mcg"),
		StartDate:  "2026-01-01",
		Note:       stringPointer("started after annual exam"),
	})
	if err != nil {
		t.Fatalf("create medication: %v", err)
	}
	med, err = api.ReplaceMedicationCourse(ctx, med.ID, client.MedicationCourseInput{
		Name:      "Levothyroxine",
		StartDate: "2026-01-02",
		Note:      stringPointer("dose held before imaging"),
	})
	if err != nil {
		t.Fatalf("replace medication: %v", err)
	}
	if med.StartDate != "2026-01-02" || med.DosageText != nil || med.Note == nil || *med.Note != "dose held before imaging" {
		t.Fatalf("medication = %#v, want replacement values", med)
	}
	meds, err := api.ListMedicationCourses(ctx, client.MedicationListOptions{Status: client.MedicationStatusAll})
	if err != nil {
		t.Fatalf("list medications: %v", err)
	}
	if len(meds) != 1 || meds[0].ID != med.ID {
		t.Fatalf("medications = %#v", meds)
	}

	slug := client.AnalyteSlugGlucose
	lab, err := api.CreateLabCollection(ctx, client.LabCollectionInput{
		CollectedAt: recordedAt,
		Note:        stringPointer("labs look stable"),
		Panels: []client.LabPanelInput{
			{
				PanelName: "Metabolic",
				Results: []client.LabResultInput{
					{
						TestName:      "Glucose",
						CanonicalSlug: &slug,
						ValueText:     "89",
						ValueNumeric:  float64Pointer(89),
						Units:         stringPointer("mg/dL"),
						Notes:         []string{"HIV 4th gen narrative", "Hep C Ab reviewed"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create lab collection: %v", err)
	}
	lab, err = api.ReplaceLabCollection(ctx, lab.ID, client.LabCollectionInput{
		CollectedAt: recordedAt.Add(24 * time.Hour),
		Note:        stringPointer("collection corrected"),
		Panels: []client.LabPanelInput{
			{
				PanelName: "Thyroid",
				Results: []client.LabResultInput{
					{
						TestName:  "TSH",
						ValueText: "3.1",
						Notes:     []string{"A1C context"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("replace lab collection: %v", err)
	}
	if lab.Panels[0].PanelName != "Thyroid" || lab.Panels[0].Results[0].TestName != "TSH" || lab.Note == nil || *lab.Note != "collection corrected" {
		t.Fatalf("lab collection = %#v, want replacement panel/result", lab)
	}
	if !slices.Equal(lab.Panels[0].Results[0].Notes, []string{"A1C context"}) {
		t.Fatalf("lab result notes = %#v", lab.Panels[0].Results[0].Notes)
	}
	labs, err := api.ListLabCollections(ctx)
	if err != nil {
		t.Fatalf("list lab collections: %v", err)
	}
	if len(labs) != 1 || labs[0].ID != lab.ID {
		t.Fatalf("lab collections = %#v", labs)
	}

	bodyUnit := client.WeightUnitLb
	body, err := api.CreateBodyComposition(ctx, client.BodyCompositionInput{
		RecordedAt:     recordedAt,
		BodyFatPercent: float64Pointer(18.7),
		WeightValue:    float64Pointer(154.2),
		WeightUnit:     &bodyUnit,
		Method:         stringPointer("smart scale"),
		Note:           stringPointer("same row as weight"),
	})
	if err != nil {
		t.Fatalf("create body composition: %v", err)
	}
	body, err = api.ReplaceBodyComposition(ctx, body.ID, client.BodyCompositionInput{
		RecordedAt:     recordedAt.Add(24 * time.Hour),
		BodyFatPercent: float64Pointer(18.1),
		Method:         stringPointer("DEXA"),
	})
	if err != nil {
		t.Fatalf("replace body composition: %v", err)
	}
	if body.BodyFatPercent == nil || *body.BodyFatPercent != 18.1 || body.WeightValue != nil || body.Method == nil || *body.Method != "DEXA" {
		t.Fatalf("body composition = %#v, want replacement values", body)
	}
	bodies, err := api.ListBodyComposition(ctx, client.BodyCompositionListOptions{Limit: 1})
	if err != nil {
		t.Fatalf("list body composition: %v", err)
	}
	if len(bodies) != 1 || bodies[0].ID != body.ID {
		t.Fatalf("body composition list = %#v", bodies)
	}

	imaging, err := api.CreateImaging(ctx, client.ImagingRecordInput{
		PerformedAt: recordedAt,
		Modality:    "X-ray",
		BodySite:    stringPointer("chest"),
		Title:       stringPointer("Chest X-ray"),
		Summary:     "No acute cardiopulmonary abnormality.",
		Impression:  stringPointer("Normal chest radiograph."),
		Note:        stringPointer("ordered for cough"),
		Notes:       []string{"XR TOE RIGHT narrative", "US Head/Neck findings"},
	})
	if err != nil {
		t.Fatalf("create imaging: %v", err)
	}
	imaging, err = api.ReplaceImaging(ctx, imaging.ID, client.ImagingRecordInput{
		PerformedAt: recordedAt.Add(24 * time.Hour),
		Modality:    "CT",
		BodySite:    stringPointer("chest"),
		Summary:     "Stable small pulmonary nodule.",
		Notes:       []string{"US abdominal findings"},
	})
	if err != nil {
		t.Fatalf("replace imaging: %v", err)
	}
	if imaging.Modality != "CT" || imaging.Title != nil || imaging.Summary != "Stable small pulmonary nodule." {
		t.Fatalf("imaging = %#v, want replacement values", imaging)
	}
	if !slices.Equal(imaging.Notes, []string{"US abdominal findings"}) {
		t.Fatalf("imaging notes = %#v", imaging.Notes)
	}
	images, err := api.ListImaging(ctx, client.ImagingListOptions{Modality: stringPointer("ct"), BodySite: stringPointer("CHEST")})
	if err != nil {
		t.Fatalf("list imaging: %v", err)
	}
	if len(images) != 1 || images[0].ID != imaging.ID {
		t.Fatalf("imaging list = %#v", images)
	}
	if err := api.DeleteBloodPressure(ctx, bp.ID); err != nil {
		t.Fatalf("delete blood pressure: %v", err)
	}
	if err := api.DeleteMedicationCourse(ctx, med.ID); err != nil {
		t.Fatalf("delete medication: %v", err)
	}
	if err := api.DeleteLabCollection(ctx, lab.ID); err != nil {
		t.Fatalf("delete lab collection: %v", err)
	}
	if err := api.DeleteBodyComposition(ctx, body.ID); err != nil {
		t.Fatalf("delete body composition: %v", err)
	}
	if err := api.DeleteImaging(ctx, imaging.ID); err != nil {
		t.Fatalf("delete imaging: %v", err)
	}
}

func intPointer(value int) *int {
	return &value
}

func stringPointer(value string) *string {
	return &value
}

func float64Pointer(value float64) *float64 {
	return &value
}
