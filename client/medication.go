package client

import (
	"context"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
)

type MedicationStatus string

const (
	MedicationStatusActive MedicationStatus = "active"
	MedicationStatusAll    MedicationStatus = "all"
)

type MedicationCourseInput struct {
	Name       string
	DosageText *string
	StartDate  string
	EndDate    *string
	Note       *string
}

type MedicationListOptions struct {
	Status MedicationStatus
}

type MedicationCourse struct {
	ID         int
	Name       string
	DosageText *string
	StartDate  string
	EndDate    *string
	Note       *string
	Source     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

func (c *LocalClient) CreateMedicationCourse(ctx context.Context, input MedicationCourseInput) (MedicationCourse, error) {
	service, err := c.localService()
	if err != nil {
		return MedicationCourse{}, err
	}
	item, err := service.CreateMedicationCourse(ctx, toHealthMedicationCourseInput(input))
	if err != nil {
		return MedicationCourse{}, err
	}
	return fromHealthMedicationCourse(item), nil
}

func (c *LocalClient) ReplaceMedicationCourse(ctx context.Context, id int, input MedicationCourseInput) (MedicationCourse, error) {
	service, err := c.localService()
	if err != nil {
		return MedicationCourse{}, err
	}
	item, err := service.ReplaceMedicationCourse(ctx, id, toHealthMedicationCourseInput(input))
	if err != nil {
		return MedicationCourse{}, err
	}
	return fromHealthMedicationCourse(item), nil
}

func (c *LocalClient) DeleteMedicationCourse(ctx context.Context, id int) error {
	service, err := c.localService()
	if err != nil {
		return err
	}
	return service.DeleteMedicationCourse(ctx, id)
}

func (c *LocalClient) ListMedicationCourses(ctx context.Context, options MedicationListOptions) ([]MedicationCourse, error) {
	service, err := c.localService()
	if err != nil {
		return nil, err
	}
	items, err := service.ListMedications(ctx, health.MedicationListParams{
		Status: health.MedicationStatus(options.Status),
	})
	if err != nil {
		return nil, err
	}
	return fromHealthMedicationCourses(items), nil
}

func toHealthMedicationCourseInput(input MedicationCourseInput) health.MedicationCourseInput {
	return health.MedicationCourseInput(input)
}

func fromHealthMedicationCourses(items []health.MedicationCourse) []MedicationCourse {
	out := make([]MedicationCourse, 0, len(items))
	for _, item := range items {
		out = append(out, fromHealthMedicationCourse(item))
	}
	return out
}

func fromHealthMedicationCourse(item health.MedicationCourse) MedicationCourse {
	return MedicationCourse{
		ID:         item.ID,
		Name:       item.Name,
		DosageText: item.DosageText,
		StartDate:  item.StartDate,
		EndDate:    item.EndDate,
		Note:       item.Note,
		Source:     item.Source,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
		DeletedAt:  item.DeletedAt,
	}
}
