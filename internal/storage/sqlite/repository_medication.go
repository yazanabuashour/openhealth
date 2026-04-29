package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite/sqlc"
)

func (r *Repository) ListMedicationCourses(ctx context.Context, status health.MedicationStatus, today string) ([]health.MedicationCourse, error) {
	items := []health.MedicationCourse{}
	switch status {
	case health.MedicationStatusActive:
		rows, err := r.queries.ListActiveMedicationCourses(ctx, &today)
		if err != nil {
			return nil, wrapDatabaseError("failed to list health medications", err)
		}
		for _, row := range rows {
			item, err := toMedicationCourseFields(row.ID, row.Name, row.DosageText, row.StartDate, row.EndDate, row.Note, row.Source, row.CreatedAt, row.UpdatedAt, row.DeletedAt)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
	case health.MedicationStatusAll:
		rows, err := r.queries.ListMedicationCourses(ctx)
		if err != nil {
			return nil, wrapDatabaseError("failed to list health medications", err)
		}
		for _, row := range rows {
			item, err := toMedicationCourseFields(row.ID, row.Name, row.DosageText, row.StartDate, row.EndDate, row.Note, row.Source, row.CreatedAt, row.UpdatedAt, row.DeletedAt)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
	default:
		return nil, &health.ValidationError{Message: "status must be 'active' or 'all'"}
	}
	return items, nil
}

func (r *Repository) CreateMedicationCourse(ctx context.Context, params health.CreateMedicationCourseParams) (health.MedicationCourse, error) {
	row, err := r.queries.CreateMedicationCourse(ctx, sqlc.CreateMedicationCourseParams{
		Name:       params.Name,
		DosageText: params.DosageText,
		StartDate:  params.StartDate,
		EndDate:    params.EndDate,
		Note:       params.Note,
		Source:     params.Source,
		CreatedAt:  serializeInstant(params.CreatedAt),
		UpdatedAt:  serializeInstant(params.UpdatedAt),
	})
	if err != nil {
		return health.MedicationCourse{}, wrapDatabaseError("failed to create health medication course", err)
	}
	return toMedicationCourseFields(row.ID, row.Name, row.DosageText, row.StartDate, row.EndDate, row.Note, row.Source, row.CreatedAt, row.UpdatedAt, row.DeletedAt)
}

func (r *Repository) UpdateMedicationCourse(ctx context.Context, params health.UpdateMedicationCourseParams) (health.MedicationCourse, error) {
	row, err := r.queries.UpdateMedicationCourse(ctx, sqlc.UpdateMedicationCourseParams{
		Name:       params.Name,
		DosageText: params.DosageText,
		StartDate:  params.StartDate,
		EndDate:    params.EndDate,
		Note:       params.Note,
		UpdatedAt:  serializeInstant(params.UpdatedAt),
		ID:         int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return health.MedicationCourse{}, &health.NotFoundError{
				Resource: "health_medication_course",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return health.MedicationCourse{}, wrapDatabaseError("failed to update health medication course", err)
	}
	return toMedicationCourseFields(row.ID, row.Name, row.DosageText, row.StartDate, row.EndDate, row.Note, row.Source, row.CreatedAt, row.UpdatedAt, row.DeletedAt)
}

func (r *Repository) DeleteMedicationCourse(ctx context.Context, params health.DeleteMedicationCourseParams) error {
	deletedAt := serializeInstant(params.DeletedAt)
	_, err := r.queries.DeleteMedicationCourse(ctx, sqlc.DeleteMedicationCourseParams{
		DeletedAt: &deletedAt,
		UpdatedAt: serializeInstant(params.UpdatedAt),
		ID:        int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &health.NotFoundError{
				Resource: "health_medication_course",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return wrapDatabaseError("failed to delete health medication course", err)
	}
	return nil
}

func (r *Repository) CountActiveMedicationCourses(ctx context.Context, today string) (int, error) {
	count, err := r.queries.CountActiveMedicationCourses(ctx, &today)
	if err != nil {
		return 0, wrapDatabaseError("failed to count active medications", err)
	}
	return int(count), nil
}

func toMedicationCourseFields(id int64, name string, dosageText *string, startDate string, endDate *string, note *string, source string, createdAtValue string, updatedAtValue string, deletedAtValue *string) (health.MedicationCourse, error) {
	createdAt, err := parseInstant(createdAtValue)
	if err != nil {
		return health.MedicationCourse{}, err
	}
	updatedAt, err := parseInstant(updatedAtValue)
	if err != nil {
		return health.MedicationCourse{}, err
	}
	deletedAt, err := parseOptionalInstant(deletedAtValue)
	if err != nil {
		return health.MedicationCourse{}, err
	}
	return health.MedicationCourse{
		ID:         int(id),
		Name:       name,
		DosageText: dosageText,
		StartDate:  startDate,
		EndDate:    endDate,
		Note:       note,
		Source:     source,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		DeletedAt:  deletedAt,
	}, nil
}
