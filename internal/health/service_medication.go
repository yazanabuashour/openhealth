package health

import (
	"context"
	"time"
)

func (s *service) ListMedications(ctx context.Context, params MedicationListParams) ([]MedicationCourse, error) {
	status, err := normalizeMedicationStatus(params.Status)
	if err != nil {
		return nil, err
	}
	return s.repo.ListMedicationCourses(ctx, status, s.now().UTC().Format(time.DateOnly))
}

func (s *service) CreateMedicationCourse(ctx context.Context, input MedicationCourseInput) (MedicationCourse, error) {
	normalized, err := normalizeMedicationCourseInput(input)
	if err != nil {
		return MedicationCourse{}, err
	}
	now := s.now().UTC()
	return s.repo.CreateMedicationCourse(ctx, CreateMedicationCourseParams{
		Name:       normalized.Name,
		DosageText: normalized.DosageText,
		StartDate:  normalized.StartDate,
		EndDate:    normalized.EndDate,
		Note:       normalized.Note,
		Source:     manualSource,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
}

func (s *service) ReplaceMedicationCourse(ctx context.Context, id int, input MedicationCourseInput) (MedicationCourse, error) {
	if err := validateRecordID(id); err != nil {
		return MedicationCourse{}, err
	}
	normalized, err := normalizeMedicationCourseInput(input)
	if err != nil {
		return MedicationCourse{}, err
	}
	return s.repo.UpdateMedicationCourse(ctx, UpdateMedicationCourseParams{
		ID:         id,
		Name:       normalized.Name,
		DosageText: normalized.DosageText,
		StartDate:  normalized.StartDate,
		EndDate:    normalized.EndDate,
		Note:       normalized.Note,
		UpdatedAt:  s.now().UTC(),
	})
}

func (s *service) DeleteMedicationCourse(ctx context.Context, id int) error {
	if err := validateRecordID(id); err != nil {
		return err
	}
	now := s.now().UTC()
	return s.repo.DeleteMedicationCourse(ctx, DeleteMedicationCourseParams{
		ID:        id,
		DeletedAt: now,
		UpdatedAt: now,
	})
}
