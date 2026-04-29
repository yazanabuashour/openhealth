package health

import "time"

type SleepEntry struct {
	ID               int
	RecordedAt       time.Time
	QualityScore     int
	WakeupCount      *int
	Note             *string
	Source           string
	SourceRecordHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type SleepWriteStatus string

const (
	SleepWriteStatusCreated       SleepWriteStatus = "created"
	SleepWriteStatusAlreadyExists SleepWriteStatus = "already_exists"
	SleepWriteStatusUpdated       SleepWriteStatus = "updated"
)

type SleepWriteResult struct {
	Entry  SleepEntry
	Status SleepWriteStatus
}

type SleepInput struct {
	RecordedAt   time.Time
	QualityScore int
	WakeupCount  *int
	Note         *string
}

type CreateSleepEntryParams struct {
	RecordedAt       time.Time
	QualityScore     int
	WakeupCount      *int
	Note             *string
	Source           string
	SourceRecordHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UpdateSleepEntryParams struct {
	ID           int
	RecordedAt   *time.Time
	QualityScore *int
	WakeupCount  *int
	Note         *string
	UpdatedAt    time.Time
}

type DeleteSleepEntryParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
}

type FindManualSleepEntryParams struct {
	RecordedAt time.Time
}
