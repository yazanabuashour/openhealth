package health_test

import (
	"context"
	"testing"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite"
	"github.com/yazanabuashour/openhealth/internal/testutil"
)

func TestServiceTargetResolution(t *testing.T) {
	t.Parallel()

	db := testutil.NewSQLiteDB(t)
	repo := sqlite.NewRepository(db)
	service := health.NewService(repo, health.WithClock(func() time.Time {
		return time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	}))
	ctx := context.Background()
	date := time.Date(2026, 3, 29, 0, 0, 0, 0, time.UTC)

	body, err := service.CreateBodyComposition(ctx, health.BodyCompositionInput{
		RecordedAt:     date,
		BodyFatPercent: float64Pointer(18.7),
	})
	if err != nil {
		t.Fatalf("create body composition: %v", err)
	}
	if _, err := service.CreateBodyComposition(ctx, health.BodyCompositionInput{
		RecordedAt:     date,
		BodyFatPercent: float64Pointer(19.1),
	}); err != nil {
		t.Fatalf("create second body composition: %v", err)
	}
	gotBody, rejection, err := service.ResolveBodyCompositionTarget(ctx, health.BodyCompositionTarget{ID: body.ID})
	if err != nil || rejection != "" || gotBody.ID != body.ID {
		t.Fatalf("body id target = %#v rejection=%q err=%v", gotBody, rejection, err)
	}
	_, rejection, err = service.ResolveBodyCompositionTarget(ctx, health.BodyCompositionTarget{RecordedAt: &date})
	if err != nil || rejection != "multiple matching body composition entries; target is ambiguous" {
		t.Fatalf("body date rejection = %q err=%v", rejection, err)
	}

	lab, err := service.CreateLabCollection(ctx, health.LabCollectionInput{
		CollectedAt: date,
		Panels: []health.LabPanelInput{{
			PanelName: "Metabolic",
			Results:   []health.LabResultInput{{TestName: "Glucose", ValueText: "89"}},
		}},
	})
	if err != nil {
		t.Fatalf("create lab: %v", err)
	}
	if _, err := service.CreateLabCollection(ctx, health.LabCollectionInput{
		CollectedAt: date,
		Panels: []health.LabPanelInput{{
			PanelName: "Thyroid",
			Results:   []health.LabResultInput{{TestName: "TSH", ValueText: "3.1"}},
		}},
	}); err != nil {
		t.Fatalf("create second lab: %v", err)
	}
	gotLab, rejection, err := service.ResolveLabCollectionTarget(ctx, health.LabCollectionTarget{ID: lab.ID})
	if err != nil || rejection != "" || gotLab.ID != lab.ID {
		t.Fatalf("lab id target = %#v rejection=%q err=%v", gotLab, rejection, err)
	}
	_, rejection, err = service.ResolveLabCollectionTarget(ctx, health.LabCollectionTarget{CollectedAt: &date})
	if err != nil || rejection != "multiple matching lab collections; target is ambiguous" {
		t.Fatalf("lab date rejection = %q err=%v", rejection, err)
	}

	med, err := service.CreateMedicationCourse(ctx, health.MedicationCourseInput{Name: "Levothyroxine", StartDate: "2026-01-01"})
	if err != nil {
		t.Fatalf("create medication: %v", err)
	}
	if _, err := service.CreateMedicationCourse(ctx, health.MedicationCourseInput{Name: "levothyroxine", StartDate: "2026-01-01"}); err != nil {
		t.Fatalf("create second medication: %v", err)
	}
	gotMed, rejection, err := service.ResolveMedicationTarget(ctx, health.MedicationTarget{ID: med.ID})
	if err != nil || rejection != "" || gotMed.ID != med.ID {
		t.Fatalf("medication id target = %#v rejection=%q err=%v", gotMed, rejection, err)
	}
	_, rejection, err = service.ResolveMedicationTarget(ctx, health.MedicationTarget{Name: "LEVOTHYROXINE", StartDate: "2026-01-01"})
	if err != nil || rejection != "multiple matching medications; target is ambiguous" {
		t.Fatalf("medication natural target rejection = %q err=%v", rejection, err)
	}

	imaging, err := service.CreateImaging(ctx, health.ImagingRecordInput{
		PerformedAt: date,
		Modality:    "MRI",
		Summary:     "Normal brain MRI",
	})
	if err != nil {
		t.Fatalf("create imaging: %v", err)
	}
	if _, err := service.CreateImaging(ctx, health.ImagingRecordInput{
		PerformedAt: date,
		Modality:    "CT",
		Summary:     "Normal head CT",
	}); err != nil {
		t.Fatalf("create second imaging: %v", err)
	}
	gotImaging, rejection, err := service.ResolveImagingTarget(ctx, health.ImagingTarget{ID: imaging.ID})
	if err != nil || rejection != "" || gotImaging.ID != imaging.ID {
		t.Fatalf("imaging id target = %#v rejection=%q err=%v", gotImaging, rejection, err)
	}
	_, rejection, err = service.ResolveImagingTarget(ctx, health.ImagingTarget{PerformedAt: &date})
	if err != nil || rejection != "multiple matching imaging records; target is ambiguous" {
		t.Fatalf("imaging date rejection = %q err=%v", rejection, err)
	}
}
