package health

import "time"

type BodyCompositionTarget struct {
	ID         int
	RecordedAt *time.Time
}

type SleepTarget struct {
	ID         int
	RecordedAt *time.Time
}

type ImagingTarget struct {
	ID          int
	PerformedAt *time.Time
}

type MedicationTarget struct {
	ID        int
	Name      string
	StartDate string
}

type LabCollectionTarget struct {
	ID          int
	CollectedAt *time.Time
}
