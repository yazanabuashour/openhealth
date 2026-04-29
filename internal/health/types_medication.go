package health

import "time"

type MedicationStatus string

const (
	MedicationStatusActive MedicationStatus = "active"
	MedicationStatusAll    MedicationStatus = "all"
)

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

type MedicationCourseInput struct {
	Name       string
	DosageText *string
	StartDate  string
	EndDate    *string
	Note       *string
}

type MedicationListParams struct {
	Status MedicationStatus
}

type CreateMedicationCourseParams struct {
	Name       string
	DosageText *string
	StartDate  string
	EndDate    *string
	Note       *string
	Source     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type UpdateMedicationCourseParams struct {
	ID         int
	Name       string
	DosageText *string
	StartDate  string
	EndDate    *string
	Note       *string
	UpdatedAt  time.Time
}

type DeleteMedicationCourseParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
}
