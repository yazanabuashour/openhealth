package health

import "time"

type BloodPressureEntry struct {
	ID               int
	RecordedAt       time.Time
	Systolic         int
	Diastolic        int
	Pulse            *int
	Note             *string
	Source           string
	SourceRecordHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type BloodPressureRecordInput struct {
	RecordedAt time.Time
	Systolic   int
	Diastolic  int
	Pulse      *int
	Note       *string
}

type CreateBloodPressureEntryParams struct {
	RecordedAt       time.Time
	Systolic         int
	Diastolic        int
	Pulse            *int
	Note             *string
	Source           string
	SourceRecordHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UpdateBloodPressureEntryParams struct {
	ID         int
	RecordedAt time.Time
	Systolic   int
	Diastolic  int
	Pulse      *int
	Note       *string
	UpdatedAt  time.Time
}

type DeleteBloodPressureEntryParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
}
